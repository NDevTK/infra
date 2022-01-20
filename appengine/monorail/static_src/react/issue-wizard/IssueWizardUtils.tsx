// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {CustomQuestion, IssueCategory} from "./IssueWizardTypes";

// this function is used to get the issue list belong to different persona
// when a user group is selected a list of related issue categories will show up
export function GetCategoriesByPersona (categories: IssueCategory[]):Map<string, string[]> {
  const categoriesByPersonaMap = new Map<string, string[]>();

  categories.forEach((category) => {
    if (category.enabled) {
      const currentIssuePersona = category.persona;
      const currentCategoryName = category.name;
      const categries = categoriesByPersonaMap.get(currentIssuePersona) ?? [];
      categries.push(currentCategoryName);
      categoriesByPersonaMap.set(currentIssuePersona, categries);
    }
  });

  return categoriesByPersonaMap;
}

// this function is used to get the customer questions belong to different issue category
// the customer question page will render base on these data
export function GetQuestionsByCategory(categories: IssueCategory[]): Map<string, CustomQuestion[] | null> {
  const questionsByCategoryMap = new Map<string, CustomQuestion[] | null>();
  categories.forEach((category) => {
    questionsByCategoryMap.set(category.name, category.customQuestions ?? null);
  })
  return questionsByCategoryMap;
}
