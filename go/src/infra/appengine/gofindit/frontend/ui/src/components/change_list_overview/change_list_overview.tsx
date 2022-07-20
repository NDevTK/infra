// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { ChangeListDetails } from './../../services/analysis_details';

interface Props {
  changeList: ChangeListDetails;
}

const ChangeListOverview = ({changeList}: Props) => {
  return (
    <table width='50%' >
      <colgroup>
        <col style={{ width: '30%' }} />
        <col style={{ width: '70%' }} />
      </colgroup>
      <thead>
        <tr>
          <td colSpan={2} >
            <a href={changeList.url} >
              {changeList.title}
            </a>
          </td>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>
            Status:
          </td>
          <td>
            {changeList.status}
          </td>
        </tr>
        <tr>
          <td>
            Submitted time:
          </td>
          <td>
            {changeList.submitTime}
          </td>
        </tr>
        <tr>
          <td>
            Commit position:
          </td>
          <td>
            {changeList.commitPosition}
          </td>
        </tr>
      </tbody>
    </table>
  );
}

export default ChangeListOverview;
