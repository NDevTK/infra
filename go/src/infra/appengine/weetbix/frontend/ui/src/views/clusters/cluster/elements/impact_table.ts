// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {
  customElement,
  html,
  LitElement,
  property,
} from 'lit-element';
import { Ref } from 'react';

import { Cluster, Counts } from '../../../../services/cluster';

const metric = (counts: Counts): string => {
  return counts.nominal || '0';
};

@customElement('impact-table')
export class ImpactTable extends LitElement {
  @property({ attribute: false })
    currentCluster!: Cluster;

  @property({ attribute: false })
    ref: Ref<ImpactTable> | null = null;

  render() {
    return html`
    <table data-testid="impact-table">
        <thead>
            <tr>
                <th></th>
                <th>1 day</th>
                <th>3 days</th>
                <th>7 days</th>
            </tr>
        </thead>
        <tbody class="data">
            <tr>
                <th>User Cls Failed Presubmit</th>
                <td class="number">${metric(this.currentCluster.userClsFailedPresubmit.oneDay)}</td>
                <td class="number">${metric(this.currentCluster.userClsFailedPresubmit.threeDay)}</td>
                <td class="number">${metric(this.currentCluster.userClsFailedPresubmit.sevenDay)}</td>
            </tr>
            <tr>
                <th>Presubmit-Blocking Failures Exonerated</th>
                <td class="number">${metric(this.currentCluster.criticalFailuresExonerated.oneDay)}</td>
                <td class="number">${metric(this.currentCluster.criticalFailuresExonerated.threeDay)}</td>
                <td class="number">${metric(this.currentCluster.criticalFailuresExonerated.sevenDay)}</td>
            </tr>
            <tr>
                <th>Total Failures</th>
                <td class="number">${metric(this.currentCluster.failures.oneDay)}</td>
                <td class="number">${metric(this.currentCluster.failures.threeDay)}</td>
                <td class="number">${metric(this.currentCluster.failures.sevenDay)}</td>
            </tr>
        </tbody>
    </table>`;
  }
}
