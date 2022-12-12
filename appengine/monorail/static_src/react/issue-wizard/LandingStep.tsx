// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import React from 'react';
import { makeStyles, withStyles } from '@material-ui/styles';
import { blue, yellow, red, grey } from '@material-ui/core/colors';
import FormControlLabel from '@material-ui/core/FormControlLabel';
import Checkbox, { CheckboxProps } from '@material-ui/core/Checkbox';
import SelectMenu from './SelectMenu.tsx';
import { RadioDescription } from './RadioDescription/RadioDescription.tsx';
import {GetCategoriesByPersona} from './IssueWizardUtils.tsx';
import {ISSUE_WIZARD_QUESTIONS} from './IssueWizardConfig.ts';
import DotMobileStepper from './DotMobileStepper.tsx';
import {IssueWizardPersona} from './IssueWizardTypes.tsx';

const CustomCheckbox = withStyles({
  root: {
    color: blue[400],
    '&$checked': {
      color: blue[600],
    },
  },
  checked: {},
})((props: CheckboxProps) => <Checkbox color="default" {...props} />);

const useStyles = makeStyles({
  pad: {
    margin: '10px, 20px',
    display: 'inline-block',
  },
  flex: {
    display: 'flex',
  },
  warningBox: {
    minHeight: '10vh',
    borderStyle: 'solid',
    borderWidth: '2px',
    borderColor: yellow[800],
    borderRadius: '8px',
    background: yellow[50],
    padding: '0px 20px 1em',
    margin: '1rem 0'
  },
  warningHeader: {
    color: yellow[800],
    fontSize: '16px',
    fontWeight: '500',
  },
  star: {
    color: red[700],
    marginRight: '8px',
    fontSize: '16px',
    display: 'inline-block',
  },
  header: {
    color: grey[900],
    fontSize: '1.5rem',
    margin: '1rem 0',
  },
  subheader: {
    color: grey[700],
    fontSize: '1.125rem',
    margin: '1rem 0',
  },
  alertDetail: {
    fontSize: '16px',
  },
  link: {
    fontSize: '20px',
    fontWeight: 'bolder',
    textDecoration: 'underline',
  },
  red: {
    color: red[600],
  },
  line: {
    color: grey[200],
    marginTop: '1.5rem',
    minWidth: '360px',
  }
});

type Props = {
  userPersona: IssueWizardPersona,
  setUserPersona: Function,
  category: string,
  setCategory: Function,
  setActiveStep: Function,
};

export default function LandingStep(props: Props) {

  const {userPersona, setUserPersona, category, setCategory, setActiveStep} = props;
  const classes = useStyles();

  const categoriesByPersonaMap = GetCategoriesByPersona(ISSUE_WIZARD_QUESTIONS);

  const [categoryList, setCategoryList] = React.useState(categoriesByPersonaMap.get(userPersona));
  const [checkExisting, setCheckExisting] = React.useState(false);

  const handleCheckChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setCheckExisting(event.target.checked);
  };

  const onSelectUserPersona = (userPersona: string) => {
    setUserPersona(userPersona);
    setCategoryList(categoriesByPersonaMap.get(userPersona));
    setCategory('');
  }

  const contributorAlert = () => {
    return (
      <div>
        <div className={classes.subheader}>
          Prefer to file an issue manually?
        </div>
        <div className={classes.alertDetail}>
          It's usually best to work through this short wizard so that your issue is given the labels needed for the right team to see it.
          Otherwise it might take longer for your issue to be triaged and resolved.
        </div>
        <div className={classes.alertDetail}>
          However, if you are a Chromium contributor and none of the other options apply, you may use the
          <a className={classes.link} href="entry"> regular issue entry form</a>.
        </div>
      </div>
    );
  }

  const nextEnabled = (userPersona != IssueWizardPersona.Contributor) && checkExisting && (category != '');
  return (
    <>
      <p className={classes.header}>Report an issue with Chromium</p>
      <p className={classes.subheader}>
        We want you to enter the best possible issue report so that the project team members
        can act on it effectively. The following steps will help route your issue to the correct
        people.
      </p>
      <p className={classes.subheader}>
        Please select your following role: <span className={classes.red}>*</span>
      </p>
      <RadioDescription selectedRadio={userPersona} onClickRadio={onSelectUserPersona} />
      { userPersona === IssueWizardPersona.Contributor ? contributorAlert() :
        <div>
          <div className={classes.subheader}>
            Which of the following best describes the issue that you are reporting? <span className={classes.red}>*</span>
          </div>
          <SelectMenu optionsList={categoryList} selectedOption={category} setOption={setCategory} />
          <div className={classes.warningBox}>
            <p className={classes.warningHeader}> <span className={classes.star}>*</span>Avoid duplicate issue reports:</p>
            <div>
              <FormControlLabel className={classes.pad}
                control={
                  <CustomCheckbox
                    checked={checkExisting}
                    onChange={handleCheckChange}
                    name="warningCheck"
                  />
                }
                label={
                  <span>By checking this box, I'm acknowledging that I have searched for <a href="/p/chromium/issues/list" target="_blank">existing issues</a> that already report this problem.</span>
                }
              />
            </div>
          </div>
        </div>
      }
      { userPersona === IssueWizardPersona.Contributor ? null :
        <DotMobileStepper nextEnabled={nextEnabled} activeStep={0} setActiveStep={setActiveStep}/>
      }
    </>
  );
}
