// TODO: create a `monorail/frontend/config/` folder to store all the feature config file
import {IssueCategory, IssueWizardUserGroup} from "./IssueWizardTypes";

export const IssueWizardMetaData:IssueCategory[] = [
  {
    name: 'UI',
    user: IssueWizardUserGroup.EndUser,
    enabled: true,
  },
  {
    name: 'Network / Downloading',
    user: IssueWizardUserGroup.EndUser,
    enabled: true,
  },
  {
    name: 'Audio / Video',
    user: IssueWizardUserGroup.EndUser,
    enabled: true,
  },
  {
    name: 'Content',
    user: IssueWizardUserGroup.EndUser,
    enabled: true,
  },
  {
    name: 'Apps',
    user: IssueWizardUserGroup.EndUser,
    enabled: true,
  },
  {
    name: 'Extensions / Themes',
    user: IssueWizardUserGroup.EndUser,
    enabled: true,
  },
  {
    name: 'Webstore',
    user: IssueWizardUserGroup.EndUser,
    enabled: true,
  },
  {
    name: 'Sync',
    user: IssueWizardUserGroup.EndUser,
    enabled: true,
  },
  {
    name: 'Enterprise',
    user: IssueWizardUserGroup.EndUser,
    enabled: true,
  },
  {
    name: 'Installation',
    user: IssueWizardUserGroup.EndUser,
    enabled: true,
  },
  {
    name: 'Crashes',
    user: IssueWizardUserGroup.EndUser,
    enabled: true,
  },
  {
    name: 'Security',
    user: IssueWizardUserGroup.EndUser,
    enabled: true,
  },
  {
    name: 'Other',
    user: IssueWizardUserGroup.EndUser,
    enabled: true,
  },
  {
    name: 'API',
    user: IssueWizardUserGroup.Developer,
    enabled: true,
  },
  {
    name: 'JavaScript',
    user: IssueWizardUserGroup.Developer,
    enabled: true,
  },
  {
    name: 'Developer Tools',
    user: IssueWizardUserGroup.Developer,
    enabled: true,
  },
];
