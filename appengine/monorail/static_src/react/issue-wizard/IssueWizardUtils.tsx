import {customQuestion, IssueCategory} from "./IssueWizardTypes";

// this function is used to get the issue list belong to different user group
// when a user group is selected a list of related issue catelog will show up
export function GetIssueCategoryMap (issues: IssueCategory[]):Map<string, string[]> {
  const issueUserNameAndTopicMap = new Map<string, string[]>();

  issues.forEach((issue) => {
    if (issue.enabled) {
      const currentIssueUser = issue.user.name;
      const currentIssueName = issue.name;
      const topics = issueUserNameAndTopicMap.get(currentIssueUser) ?? [];
      topics.push(currentIssueName);
      issueUserNameAndTopicMap.set(currentIssueUser, topics);
    }
  });

  return issueUserNameAndTopicMap;
}

// this function is used to get the customer questions belong to different issue category
// the customer question page will render base on these data
export function GetIssueCustomQuestionsMap(issues: IssueCategory[]):Map<string,customQuestion[] | null> {
  const issueCustomQuestionsMap = new Map<string,customQuestion[] | null>();
  issues.forEach((issue) => {
    issueCustomQuestionsMap.set(issue.name, issue.customQuestions ?? null);
  })
  return issueCustomQuestionsMap;
}
