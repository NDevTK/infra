// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import React from 'react';
import OutlinedInput from "@material-ui/core/OutlinedInput";
import {makeStyles} from '@material-ui/styles';

const userStyles = makeStyles({
  head: {
    marginTop: '1.5rem',
    fontSize: '1rem'
  },
  inputArea: {
    width: '100%',
  },
  tip: {
    margin: '0.5rem 0'
  },
});

type Props = {
  question: string,
  tip?: string,
  updateAnswers: Function,
}

export default function CustomQuestionTextarea(props: Props): React.ReactElement {

  const classes = userStyles();

  const {question, updateAnswers, tip} = props;
  const [answer, setAnswer] = React.useState('');
  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setAnswer(e.target.value);
    updateAnswers(e.target.value);
  };

  const getQuestionInnerHtml = ()=> {
    return {__html: question};
  }

  const getTipInnerHtml = ()=> {
    return {__html: tip};
  }
  return (
    <>
      <h3 dangerouslySetInnerHTML={getQuestionInnerHtml()} className={classes.head}/>
      {tip? <div dangerouslySetInnerHTML={getTipInnerHtml()} className={classes.tip}/> : null}
      <OutlinedInput
        multiline={true}
        rows={3}
        value={answer}
        onChange={handleChange}
        className={classes.inputArea}
      />
    </>
  );
}
