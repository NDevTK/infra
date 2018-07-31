package task

import "testing"

func TestLabelSetEquals(t *testing.T) {
	a := LabelSet([]string{"label1"})
	if !a.Equals(a) {
		t.Errorf("set not equal to itself")
	}
	b := LabelSet([]string{"label1"})
	if !a.Equals(b) || !b.Equals(a) {
		t.Errorf("equivalent single element sets not equal")
	}
	c := LabelSet{}
	d := LabelSet{}
	if !c.Equals(d) {
		t.Errorf("empty sets no equal")
	}
	if c.Equals(b) {
		t.Errorf("inequal sets appear equal")
	}
	e := LabelSet([]string{"label2"})
	if a.Equals(e) {
		t.Errorf("inequivalent sets are equal")
	}
	g := LabelSet([]string{"label1", "label2"})
	h := LabelSet([]string{"label2", "label1"})
	i := LabelSet([]string{"label1", "label3"})
	if !g.Equals(h) {
		t.Errorf("equivalent 2-length sets not equal")
	}
	if h.Equals(i) {
		t.Errorf("inequivalent 2-length sets are equal")
	}
}
