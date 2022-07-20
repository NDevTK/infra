// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

interface AnalysisSummary {
  id: number;
  status: string;
  failureType: string;
  buildbucketId: number;
  builder: string;
  suspectRange: string[];
  bugs: string[];
}

interface Props {
  analysis: AnalysisSummary;
}

function processSuspectRange(suspects: string[]) {
  var suspectRangeText, suspectRangeUrl = '';

  var suspectRange = [];

  const suspectCount = suspects.length;
  if (suspectCount > 0) {
    suspectRange.push(suspects[0]);
  }
  if (suspectCount > 1) {
    suspectRange.push(suspects[suspectCount - 1]);
  }

  if (suspectRange.length > 0) {
    suspectRangeText = '[' + suspectRange.join(' ... ') + ']';
    suspectRangeUrl = `/placeholder/url?earliest=${suspectRange[0]}&latest=${suspectRange[suspectRange.length - 1]}`;
  }

  return [suspectRangeText, suspectRangeUrl];
}

const AnalysisOverview = ({analysis}: Props) => {
  const [suspectRangeText, suspectRangeUrl] = processSuspectRange(analysis.suspectRange);

  return (
    <table width='100%'>
      <colgroup>
        <col style={{ width: '15%' }} />
        <col style={{ width: '35%' }} />
        <col style={{ width: '15%' }} />
        <col style={{ width: '35%' }} />
      </colgroup>
      <tbody>
        <tr>
          <td>
            Analysis ID:
          </td>
          <td>
            {analysis.id}
          </td>
          <td>
            Buildbucket ID:
          </td>
          <td>
            <a href={`${analysis.buildbucketId}`}>
              {analysis.buildbucketId}
            </a>
          </td>
        </tr>
        <tr>
          <td>
            Status:
          </td>
          <td>
            {analysis.status}
          </td>
          <td>
            Builder:
          </td>
          <td>
            {analysis.builder}
          </td>
        </tr>
        <tr>
          <td>
            Suspect range:
          </td>
          <td>
            <a href={suspectRangeUrl}>
              {suspectRangeText}
            </a>
          </td>
          <td>
            Failure Type:
          </td>
          <td>
            {analysis.failureType}
          </td>
        </tr>
        <tr>
          <td>
            <br />
          </td>
        </tr>
        <tr>
          <td>
            Related bugs:
            <ul>
              {
                analysis.bugs.map((bugUrl) => (
                  <li key={bugUrl}>
                    <a href={bugUrl}>
                      {bugUrl}
                    </a>
                  </li>
                ))
              }
            </ul>
          </td>
        </tr>
      </tbody>
    </table>
  );
}

export default AnalysisOverview;
