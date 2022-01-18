export enum IssueWizardUser {
  EndUser = 'End User',
  Developer = 'Web Developer'
};

export enum CustomQuestionType {
  NUL, // this is used to define there is no questions
  Text,
  Input,
  Select,
}
export type customQuestions = {
  type: CustomQuestionType,
  question: string,
  answers?: string[],
  subQuestions?: customQuestions[],
};

export type IssueType = {
  name: string,
  user: IssueWizardUser,
  enabled: boolean,
  customQuestions?: customQuestions[],
};
