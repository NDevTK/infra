// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package aip

import (
	"fmt"
	"strconv"
	"strings"

	spanutil "infra/appengine/weetbix/internal/span"
)

// whereClause constructs Standard SQL WHERE clause parts from
// column definitions and a parsed AIP-160 filter.
type whereClause struct {
	table         *Table
	parameters    []QueryParameter
	namePrefix    string
	nextValueName int
}

// QueryParameter represents a query parameter.
type QueryParameter struct {
	Name  string
	Value string
}

// WhereClause creates a Standard SQL WHERE clause fragment for the given filter.
//
// The fragment will be enclosed in parentheses and does not include the "WHERE" keyword.
// For example: (column LIKE @param1)
// Also returns the query parameters which need to be given to the database.
//
// All field names are replaced with the safe database column names from the specified table.
// All user input strings are passed via query parameters, so the returned query is SQL injection safe.
func (t *Table) WhereClause(filter *Filter, parameterPrefix string) (string, []QueryParameter, error) {
	if filter.Expression == nil {
		return "(TRUE)", []QueryParameter{}, nil
	}

	q := &whereClause{
		table:      t,
		namePrefix: parameterPrefix,
	}

	clause, err := q.expressionQuery(filter.Expression)
	if err != nil {
		return "", []QueryParameter{}, err
	}
	return clause, q.parameters, nil
}

// expressionQuery returns the SQL expression equivalent to the given
// filter expression.
// An expression is a conjunction (AND) of sequences or a simple
// sequence.
//
// The returned string is an injection-safe SQL expression.
func (w *whereClause) expressionQuery(expression *Expression) (string, error) {
	factors := []string{}
	// Both Sequence and Factor is equivalent to AND of the
	// component Sequences and Factors (respectively), as we implement
	// exact match semantics and do not support ranking
	// based on the number of factors that match.
	for _, sequence := range expression.Sequences {
		for _, factor := range sequence.Factors {
			f, err := w.factorQuery(factor)
			if err != nil {
				return "", err
			}
			factors = append(factors, f)
		}
	}
	if len(factors) == 1 {
		return factors[0], nil
	}
	return "(" + strings.Join(factors, " AND ") + ")", nil
}

// factorQuery returns the SQL expression equivalent to the given
// factor. A factor is a disjunction (OR) of terms or a simple term.
//
// The returned string is an injection-safe SQL expression.
func (w *whereClause) factorQuery(factor *Factor) (string, error) {
	terms := []string{}
	for _, term := range factor.Terms {
		tq, err := w.termQuery(term)
		if err != nil {
			return "", err
		}
		terms = append(terms, tq)
	}
	if len(terms) == 1 {
		return terms[0], nil
	}
	return "(" + strings.Join(terms, " OR ") + ")", nil
}

// termQuery returns the SQL expression equivalent to the given
// term.
//
// The returned string is an injection-safe SQL expression.
func (w *whereClause) termQuery(term *Term) (string, error) {
	simpleQuery, err := w.simpleQuery(term.Simple)
	if err != nil {
		return "", err
	}
	if term.Negated {
		return fmt.Sprintf("(NOT %s)", simpleQuery), nil
	}
	return simpleQuery, nil
}

// simpleQuery returns the SQL expression equivalent to the given simple
// filter.
// The returned string is an injection-safe SQL expression.
func (w *whereClause) simpleQuery(simple *Simple) (string, error) {
	if simple.Restriction != nil {
		return w.restrictionQuery(simple.Restriction)
	} else if simple.Composite != nil {
		return w.expressionQuery(simple.Composite)
	} else {
		return "", fmt.Errorf("invalid 'simple' clause in query filter")
	}
}

// restrictionQuery returns the SQL expression equivalent to the given
// restriction.
// The returned string is an injection-safe SQL expression.
func (w *whereClause) restrictionQuery(restriction *Restriction) (string, error) {
	if restriction.Comparable.Member == nil {
		return "", fmt.Errorf("invalid comparable")
	}
	if len(restriction.Comparable.Member.Fields) > 0 {
		return "", fmt.Errorf("fields not implemented yet")
	}
	if restriction.Comparator == "" {
		arg, err := w.likeComparableValue(restriction.Comparable)
		if err != nil {
			return "", err
		}
		clauses := []string{}
		// This is a value that should be substring matched against columns
		// marked for implicit matching.
		for _, column := range w.table.columns {
			if column.implicitFilter {
				clauses = append(clauses, fmt.Sprintf("%s LIKE %s", column.databaseName, arg))
			}
		}
		return "(" + strings.Join(clauses, " OR ") + ")", nil
	} else if restriction.Comparator == "=" {
		arg, err := w.argValue(restriction.Arg)
		if err != nil {
			return "", err
		}
		column, err := w.table.FilterableColumnByName(restriction.Comparable.Member.Value)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s = %s)", column.databaseName, arg), nil
	} else if restriction.Comparator == "!=" {
		arg, err := w.argValue(restriction.Arg)
		if err != nil {
			return "", err
		}
		column, err := w.table.FilterableColumnByName(restriction.Comparable.Member.Value)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s <> %s)", column.databaseName, arg), nil
	} else if restriction.Comparator == ":" {
		arg, err := w.likeArgValue(restriction.Arg)
		if err != nil {
			return "", err
		}
		column, err := w.table.FilterableColumnByName(restriction.Comparable.Member.Value)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s LIKE %s)", column.databaseName, arg), nil
	} else {
		return "", fmt.Errorf("comparator operator not implemented yet")
	}
}

// argValue returns a SQL expression representing the value of the specified
// arg.
// The returned string is an injection-safe SQL expression.
func (w *whereClause) argValue(arg *Arg) (string, error) {
	if arg.Composite != nil {
		return "", fmt.Errorf("composite expressions in arguments not implemented yet")
	}
	if arg.Comparable == nil {
		return "", fmt.Errorf("missing comparable in argument")
	}
	return w.comparableValue(arg.Comparable)
}

// argValue returns a SQL expression representing the value of the specified
// comparable.
// The returned string is an injection-safe SQL expression.
func (w *whereClause) comparableValue(comparable *Comparable) (string, error) {
	if comparable.Member == nil {
		return "", fmt.Errorf("invalid comparable")
	}
	if len(comparable.Member.Fields) > 0 {
		return "", fmt.Errorf("fields not implemented yet")
	}
	// Bind unsanitised user input to a parameter to protect against SQL injection.
	return w.bind(comparable.Member.Value), nil

}

// likeArgValue returns a SQL expression that, when passed to the
// right hand side of a LIKE operator, performs substring matching against
// the value of the argument.
// The returned string is an injection-safe SQL expression.
func (w *whereClause) likeArgValue(arg *Arg) (string, error) {
	if arg.Composite != nil {
		return "", fmt.Errorf("composite expressions are not allowed as RHS to has (:) operator")
	}
	if arg.Comparable == nil {
		return "", fmt.Errorf("missing comparable in argument")
	}
	return w.likeComparableValue(arg.Comparable)
}

// likeComparableValue returns a SQL expression that, when passed to the
// right hand side of a LIKE operator, performs substring matching against
// the value of the comparable.
// The returned string is an injection-safe SQL expression.
func (w *whereClause) likeComparableValue(comparable *Comparable) (string, error) {
	if comparable.Member == nil {
		return "", fmt.Errorf("invalid comparable")
	}
	if len(comparable.Member.Fields) > 0 {
		return "", fmt.Errorf("fields are not allowed on the RHS of has (:) operator")
	}
	// Bind unsanitised user input to a parameter to protect against SQL injection.
	return w.bind("%" + spanutil.QuoteLike(comparable.Member.Value) + "%"), nil
}

// bind binds a new query parameter with the given value, and returns
// the name of the parameter (including '@').
// The returned string is an injection-safe SQL expression.
func (q *whereClause) bind(value string) string {
	name := q.namePrefix + strconv.Itoa(q.nextValueName)
	q.nextValueName += 1
	q.parameters = append(q.parameters, QueryParameter{Name: name, Value: value})
	return "@" + name
}
