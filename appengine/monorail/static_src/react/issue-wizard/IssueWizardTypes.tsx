export const IssueWizardUserGroup  = Object.freeze({
  EndUser: {
    name: 'End User',
    description: 'I am a user trying to do something on a website.',
  },
  Developer: {
    name: 'Web Developer',
    description: 'I am a web developer trying to build something.',
  },
  CONTRIBUTOR: {
    name: 'Chromium Contributor',
    description: 'I know about a problem in specific tests or code.',
  }
});

export type IssueWizardUserGroupType = typeof IssueWizardUserGroup [keyof typeof IssueWizardUserGroup];

export enum CustomQuestionType {
  EMPTY, // this is used to define there is no subquestions
  Text,
  Input,
  Select,
}
export type customQuestion = {
  type: CustomQuestionType,
  question: string,
  options?: string[],
  subQuestions?: customQuestion[],
};

export type IssueCategory = {
  name: string,
  user: IssueWizardUserGroupType,
  enabled: boolean,
  customQuestions?: customQuestion[],
};
