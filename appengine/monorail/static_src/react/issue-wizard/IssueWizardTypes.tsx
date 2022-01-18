export enum IssueWizardUserGroup {
  EndUser = 'End User',
  Developer = 'Web Developer'
};

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
  user: IssueWizardUserGroup,
  enabled: boolean,
  customQuestions?: customQuestion[],
};
