// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {LitElement, html, css} from 'lit-element';
import {prpcClient} from 'prpc-client-instance.js';
import {linearRegression} from 'shared/math.js';
import './chops-chart.js';

const DEFAULT_NUM_DAYS = 90;
const SECONDS_IN_DAY = 24 * 60 * 60;
const MAX_QUERY_SIZE = 90;
const MAX_DISPLAY_LINES = 10;
const predRangeType = Object.freeze({
  NEXT_MONTH: 0,
  NEXT_QUARTER: 1,
  NEXT_50: 2,
  HIDE: 3,
});
const CHART_OPTIONS = {
  animation: false,
  responsive: true,
  title: {
    display: true,
    text: 'Issues over time',
  },
  tooltips: {
    mode: 'x',
    intersect: false,
  },
  hover: {
    mode: 'x',
    intersect: false,
  },
  legend: {
    display: true,
    labels: {
      boxWidth: 15,
    },
  },
  scales: {
    xAxes: [{
      display: true,
      type: 'time',
      time:{
        parser: 'MM/DD/YYYY',
        tooltipFormat: 'll'
      },
      scaleLabel: {
        display: true,
        labelString: 'Day',
      },
    }],
    yAxes: [{
      display: true,
      ticks: {
        beginAtZero: true,
      },
      scaleLabel: {
        display: true,
        labelString: 'Value',
      },
    }],
  },
};
const COLOR_CHOICES = ['#00838F', '#B71C1C', '#2E7D32', '#00659C',
  '#5D4037', '#558B2F', '#FF6F00', '#6A1B9A', '#880E4F', '#827717'];
const BG_COLOR_CHOICES = ['#B2EBF2', '#EF9A9A', '#C8E6C9', '#B2DFDB',
  '#D7CCC8', '#DCEDC8', '#FFECB3', '#E1BEE7', '#F8BBD0', '#E6EE9C'];

export default class MrChart extends LitElement {
  /** @override */
  static get properties() {
    return {
      progress: {type: Number},
      projectName: {type: String},
      indices: {type: Array},
      values: {type: Array},
      unsupportedFields: {type: Array},
      dateRangeNotLegal: {type: Boolean},
      dateRange: {type: Number},
      frequency: {type: Number},
    };
  }

  /** @override */
  static get styles() {
    return css`
      :host {
        display: block;
        max-width: 800px;
        margin: 0 auto;
      }
      chops-chart {
        max-width: 100%;
      }
      div#options {
        max-width: 720px;
        margin: 2em auto;
        text-align: center;
      }
      div#options #unsupported-fields {
        font-weight: bold;
        color: orange;
      }
      div.align {
        display: flex;
      }
      div.align #frequency, div.align #groupBy {
        display: inline-block;
        width: 40%;
      }
      div.align #frequency #two-toggle {
        font-size: 95%;
        text-align: center;
        margin-bottom: 5px;
      }
      div.align #time, div.align #prediction {
        display: inline-block;
        width: 60%;
      }
      #dropdown {
        height: 50%;
      }
      div.section {
        display: inline-block;
        text-align: center;
      }
      div.section.input {
        padding: 4px 10px;
      }
      .menu {
        min-width: 50%;
        text-align: left;
        font-size: 12px;
        box-sizing: border-box;
        text-decoration: none;
        white-space: nowrap;
        padding: 0.25em 8px;
        transition: 0.2s background ease-in-out;
        cursor: pointer;
        color: var(--chops-link-color);
      }
      .menu:hover {
        background: hsl(0, 0%, 90%);
      }
      .choice.transparent {
        background: white;
        border-color: var(--chops-choice-color);
        border-radius: 4px;
      }
      .choice.shown {
        background: var(--chops-blue-50);
      }
      .choice {
        padding: 4px 10px;
        background: var(--chops-choice-bg);
        color: var(--chops-choice-color);
        text-decoration: none;
        display: inline-block;
      }
      .choice.checked {
        background: var(--chops-blue-50);
      }
      p .warning-message {
        display: none;
        font-size: 1.25em;
        padding: 0.25em;
        background-color: var(--chops-orange-50);
      }
      progress {
        background-color: white;
        border: 1px solid var(--chops-gray-500);
        margin: 0 0 1em;
        width: 100%;
        visibility: visible;
      }
      ::-webkit-progress-bar {
        background-color: white;
      }
      progress::-webkit-progress-value {
        transition: width 1s;
        background-color: #00838F;
      }
    `;
  }

  /** @override */
  render() {
    const doneLoading = this.progress === 1;
    return html`
      <chops-chart
        type="line"
        .options=${CHART_OPTIONS}
        .data=${this._chartData(this.indices, this.values)}
      ></chops-chart>
      <div id="options">
        <p id="unsupported-fields">
          ${this.unsupportedFields.length ? `
            Unsupported fields: ${this.unsupportedFields.join(', ')}`: ''}
        </p>
        <progress
          value=${this.progress}
          ?hidden=${doneLoading}
        >Loading chart...</progress>
        <p class="warning-message" ?hidden=${!this.searchLimitReached}>
          Note: Some results are not being counted.
          Please narrow your query.
        </p>
        <p class="warning-message" ?hidden=${!this.maxQuerySizeReached}>
          Your query is too long.
          Showing ${MAX_QUERY_SIZE} weeks from end date.
        </p>
        <p class="warning-message" ?hidden=${!this.dateRangeNotLegal}>
          Your requested date range does not exist.
          Showing ${MAX_QUERY_SIZE} days from end date.
        </p>
        <p class="warning-message" ?hidden=${!this.cannedQueryOpen}>
          Your query scope prevents closed issues from showing.
        </p>
        <div class="align">
          <div id="frequency">
            <label for="two-toggle">Choose date range:</label>
            <div id="two-toggle">
              <chops-button @click="${this._setDateRange.bind(this, 180)}"
                class="${this.dateRange === 180 ? 'choice checked': 'choice'}">
                180 Days
              </chops-button>
              <chops-button @click="${this._setDateRange.bind(this, 90)}"
                class="${this.dateRange === 90 ? 'choice checked': 'choice'}">
                90 Days
              </chops-button>
              <chops-button @click="${this._setDateRange.bind(this, 30)}"
                class="${this.dateRange === 30 ? 'choice checked': 'choice'}">
                30 Days
              </chops-button>
            </div>
          </div>
          <div id="time">
            <label for="start-date">Choose start and end date:</label>
            <br />
            <input
              type="date"
              id="start-date"
              name="start-date"
              .value=${this.startDate && this.startDate.toISOString().substr(0, 10)}
              ?disabled=${!doneLoading}
              @change=${(e) => this.startDate = MrChart.dateStringToDate(e.target.value)}
            />
            <input
              type="date"
              id="end-date"
              name="end-date"
              .value=${this.endDate && this.endDate.toISOString().substr(0, 10)}
              ?disabled=${!doneLoading}
              @change=${(e) => this.endDate = MrChart.dateStringToDate(e.target.value)}
            />
            <chops-button @click="${this._onDateChanged}" class=choice>
              Apply
            </chops-button>
          </div>
        </div>
        <div class="align">
          <div id="prediction">
            <label for="two-toggle">Choose prediction range:</label>
            <div id="two-toggle">
              <chops-button @click="${() => {this.predRange = predRangeType.NEXT_MONTH; this._fetchData()}}"
                class="${this.predRange === predRangeType.NEXT_MONTH ? 'choice checked': 'choice'}">
                Future Month
              </chops-button>
              <chops-button @click="${() => {this.predRange = predRangeType.NEXT_QUARTER; this._fetchData()}}"
                class="${this.predRange === predRangeType.NEXT_QUARTER ? 'choice checked': 'choice'}">
                Future Quarter
              </chops-button>
              <chops-button @click="${() => {this.predRange = predRangeType.NEXT_50; this._fetchData()}}"
                class="${this.predRange === predRangeType.NEXT_50 ? 'choice checked': 'choice'}">
                Future 50%
              </chops-button>
              <chops-button @click="${() => {this.predRange = predRangeType.HIDE; this._fetchData()}}"
                class="${this.predRange === predRangeType.HIDE ? 'choice checked': 'choice'}">
                Hide
              </chops-button>
            </div>
          </div>
          <div id="groupBy">
            <label for="dropdown">Choose group by:</label>
            <mr-dropdown
              id="dropdown"
              ?disabled=${!doneLoading}
              .text=${this.groupBy.display}
            >
              ${this.dropdownHTML}
            </mr-dropdown>
          </div>
        </div>
      </div>
    `;
  }

  /** @override */
  constructor() {
    super();
    this.progress = 0.05;
    this.values = [];
    this.indices = [];
    this.unsupportedFields = [];
    this.endDate = MrChart.getEndDate();
    this.startDate = MrChart.getStartDate(this.endDate, DEFAULT_NUM_DAYS);
    this.predRange = predRangeType.HIDE;
    this.groupBy = MrChart.getGroupByURL();
  }

  // Set dropdown options menu in HTML.
  async _constructDropdownMenu() {
    let response = await this._getLabelPrefixes();
    let dropdownOptions = ['None', 'Component', 'Is open', 'Status', 'Owner'];
    dropdownOptions = dropdownOptions.concat(response);
    const dropdownHTML = dropdownOptions.map((str) => html`
      <option class='menu' @click=${this._setGroupBy}>
        ${str}</option>`);
    this.dropdownHTML = html`${dropdownHTML}`;
  }

  async connectedCallback() {
    super.connectedCallback();

    if (!this.projectName || !this.projectName.length) {
      throw new Error('Attribute `projectName` required.');
    }

    // Load Chart.js before chops-chart to allow data points to render as soon as
    // they are loaded.
    await import(/* webpackChunkName: "chartjs" */ 'chart.js/dist/Chart.bundle.min.js');

    this.dispatchEvent(new Event('chartLoaded'));
    this._constructDropdownMenu();
    this._fetchData();
  }

  // Fetch corresponding data when start date or end date changes.
  _onDateChanged() {
    this._fetchData();
    const urlParams = MrChart.getSearchParams();

    // TODO(zhangtiff): Integrate with frontend routing once charts is part of the SPA.
    urlParams.set('start-date', this.startDate.toISOString().substr(0, 10));
    urlParams.set('end-date', this.endDate.toISOString().substr(0, 10));
    const newUrl = `${location.protocol}//${location.host}${location.pathname}?${urlParams.toString()}`;
    window.history.pushState({}, '', newUrl);
  }

  async _fetchData() {
    this.dateRange = Math.ceil((this.endDate - this.startDate) / (1000 * SECONDS_IN_DAY));

    // Coordinate different parameters and flags, protection against illegal queries.
    // Case for start date greater than end date.
    if (this.dateRange <= 0) {
      this.frequency = 7;
      this.dateRangeNotLegal = true;
      this.maxQuerySizeReached = false;
      this.dateRange = MAX_QUERY_SIZE;
    } else {
      this.dateRangeNotLegal = false;
      if (this.dateRange >= MAX_QUERY_SIZE * 7) {
        // Case for date range too long, requires >= MAX_QUERY_SIZE queries.
        this.frequency = 7;
        this.maxQuerySizeReached = true;
        this.dateRange = MAX_QUERY_SIZE * 7;
      } else {
        this.maxQuerySizeReached = false;
        if (this.dateRange < MAX_QUERY_SIZE) {
          // Case for small date range, displayed in daily frequency.
          this.frequency = 1;
        } else {
          // Case for medium date range, displayed in weekly frequency.
          this.frequency = 7;
        }
      }
    }
    // Set canned query flag.
    this.cannedQueryOpen = (MrChart.getSearchParams().get('can') === '2'
      && this.groupBy.value === 'open');

    // Reset chart variables except indices.
    this.progress = 0.05;

    let numTimestampsLoaded = 0;
    const timestampsChronological = MrChart.makeTimestamps(this.endDate,
      this.frequency, this.dateRange);
    const tsToIndexMap = new Map(timestampsChronological.map((ts, idx) => (
      [ts, idx]
    )));
    this.indices = MrChart.makeIndices(timestampsChronological);
    const timestamps = MrChart.sortInBisectOrder(timestampsChronological);
    this.values = new Array(timestamps.length).fill(undefined);

    const fetchPromises = timestamps.map(async (ts) => {
      const data = await this._fetchDataAtTimestamp(ts);
      const index = tsToIndexMap.get(ts);
      this.values[index] = data.issues;
      numTimestampsLoaded += 1;
      const progressValue = numTimestampsLoaded / timestamps.length;
      this.progress = progressValue;

      return data;
    });

    const chartData = await Promise.all(fetchPromises);

    this.dispatchEvent(new Event('allDataLoaded'));

    // Check if the query includes any field values that are not supported.
    const flatUnsupportedFields = chartData.reduce((acc, datum) => {
      if (datum.unsupportedField) {
        acc = acc.concat(datum.unsupportedField);
      }
      return acc;
    }, []);
    this.unsupportedFields = Array.from(new Set(flatUnsupportedFields));

    this.searchLimitReached = chartData.some((d) => d.searchLimitReached);
  }

  async _fetchDataAtTimestamp(timestamp) {
    const params = MrChart.getSearchParams();
    const query = params.get('q');
    const cannedQuery = params.get('can');
    let message = {
      timestamp: timestamp,
      projectName: this.projectName,
      query: query,
      cannedQuery: cannedQuery,
    };
    if (this.groupBy.value !== '') {
      message['groupBy'] = this.groupBy.value;
      if (this.groupBy.value === 'label') {
        message['labelPrefix'] = this.groupBy.labelPrefix;
      }
    }
    const response = await prpcClient.call('monorail.Issues',
      'IssueSnapshot', message);

    let issues;
    if (response.snapshotCount) {
      issues = response.snapshotCount.reduce((map, curr) => {
        if (curr.dimension !== undefined) {
          if (this.groupBy.value === '') {
            map.set('Issue Count', curr.count);
          } else {
            map.set(curr.dimension, curr.count);
          }
        }
        return map;
      }, new Map());
    } else {
      issues = new Map();
    }
    return {
      date: timestamp * 1000,
      issues: issues,
      unsupportedField: response.unsupportedField,
      searchLimitReached: response.searchLimitReached,
    };
  }

  // Get prefixes from the set of labels.
  async _getLabelPrefixes() {
    const projectRequestMessage = {
      project_name: this.projectName};
    const labelsResponse = await prpcClient.call(
      'monorail.Projects', 'GetLabelOptions', projectRequestMessage);
    const labelPrefixes = new Set();
    for (let i = 0; i < labelsResponse.labelOptions.length; i++) {
      let label = labelsResponse.labelOptions[i].label;
      if (label.includes('-')) {
        labelPrefixes.add(label.split('-')[0]);
      }
    }
    return Array.from(labelPrefixes);
  }

  _chartData(indices, values) {
    // Generate a map of each data line with type {dimension:string, value:array}.
    let mapValues = new Map();
    for (let i = 0; i < values.length; i++) {
      if (values[i] !== undefined) {
        values[i].forEach((value, key, map) => {
          mapValues.set(key, []);}
        );
      }
    }
    // Count the number of 0 or undefined data points.
    let count = 0;
    for (let i = 0; i < values.length; i++) {
      if (values[i] !== undefined) {
        if (values[i].size === 0) {
          count++;
        }
        // Set none-existing data points 0.
        mapValues.forEach((value, key, map) => {
          mapValues.set(key, value.concat([values[i].get(key) || 0]));
        });
      } else {
        count++;
      }
    }
    // Legend display set back to default.
    CHART_OPTIONS.legend.display = true;
    // Check if any positive valued data exist, if not, draw an array of zeros.
    if (count === values.length) {
      return {
        type: 'line',
        labels: indices,
        datasets: [{
          label: this.groupBy.labelPrefix,
          data: Array(indices.length).fill(0),
          backgroundColor: COLOR_CHOICES[0],
          borderColor: COLOR_CHOICES[0],
          showLine: true,
          fill: false,
        }],
      };
    }
    // Convert map to a dataset of lines.
    let arrayValues = [];
    mapValues.forEach((value, key, map) => {
      arrayValues.push({
        label: key,
        data: value,
        backgroundColor: COLOR_CHOICES[arrayValues.length % COLOR_CHOICES.length],
        borderColor: COLOR_CHOICES[arrayValues.length % COLOR_CHOICES.length],
        showLine: true,
        fill: false,
      });
    });
    arrayValues = MrChart.getSortedLines(arrayValues, MAX_DISPLAY_LINES);
    if (this.predRange === predRangeType.HIDE) {
      return {
        type: 'line',
        labels: indices,
        datasets: arrayValues,
      };
    }

    let predictedValues = [];
    let originalData, predictedData, maxData, minData;
    let currColor;
    let currBGColor;
    // Check if displayed values > MAX_DISPLAY_LINES, hide legend.
    if (arrayValues.length * 4 > MAX_DISPLAY_LINES) {
      CHART_OPTIONS.legend.display = false;
    } else {
      CHART_OPTIONS.legend.display = true;
    }
    for (let i = 0; i < arrayValues.length; i++) {
      [originalData, predictedData, maxData, minData] =
        MrChart.getAllData(indices, arrayValues[i]['data'], this.dateRange,
          this.predRange, this.frequency);
      currColor = COLOR_CHOICES[i % COLOR_CHOICES.length];
      currBGColor = BG_COLOR_CHOICES[i % COLOR_CHOICES.length];
      predictedValues = predictedValues.concat([{
        label: arrayValues[i]['label'],
        backgroundColor: currColor,
        borderColor: currColor,
        data: originalData,
        showLine: true,
        fill: false,
      },{
        label: arrayValues[i]['label'].concat(' prediction'),
        backgroundColor: currColor,
        borderColor: currColor,
        borderDash: [5, 5],
        data: predictedData,
        pointRadius: 0,
        showLine: true,
        fill: false,
      },{
        label: arrayValues[i]['label'].concat(' lower error'),
        backgroundColor: currBGColor,
        borderColor: currBGColor,
        borderDash: [5, 5],
        data: minData,
        pointRadius: 0,
        showLine: true,
        hidden: true,
        fill: false,
      },{
        label: arrayValues[i]['label'].concat(' upper error'),
        backgroundColor: currBGColor,
        borderColor: currBGColor,
        borderDash: [5, 5],
        data: maxData,
        pointRadius: 0,
        showLine: true,
        hidden: true,
        fill: '-1',
      }]);
    }
    return {
      type: 'scatter',
      datasets: predictedValues,
    };
  }

  // Change group by based on dropdown menu selection.
  _setGroupBy(e) {
    switch(e.target.text) {
      case 'None':
        this.groupBy = {value: ''};
        break;
      case 'Is open':
        this.groupBy = {value: 'open'};
        break;
      case 'Owner':
      case 'Component':
      case 'Status':
        this.groupBy = {value: e.target.text.toLowerCase()};
        break;
      default:
        this.groupBy = {value: 'label', labelPrefix: e.target.text};
    }
    this.groupBy['display'] = e.target.text;
    this.shadowRoot.querySelector('#dropdown').text = e.target.text;
    this.shadowRoot.querySelector('#dropdown').close();
    this._fetchData();
    // Set groupby URL params.
    const urlParams = MrChart.getSearchParams();
    urlParams.set('groupby', this.groupBy.value);
    if (this.groupBy.value === 'label') {
      urlParams.set('labelPrefix', this.groupBy.labelPrefix);
    } else {
      urlParams.set('labelPrefix', '');
    }
    const newUrl = `${location.protocol}//${location.host}${location.pathname}?${urlParams.toString()}`;
    window.history.pushState({}, '', newUrl);
  }

  // Change date range and frequency based on button clicked.
  _setDateRange(dateRange) {
    if (this.dateRange !== dateRange) {
      this.startDate = new Date(this.endDate.getTime() - 1000 * SECONDS_IN_DAY * dateRange);
      this._onDateChanged();
      window.getTSMonClient().recordDateRangeChange(dateRange);
    }
  }

  // Move first, last, and median to the beginning of the array, recursively.
  static sortInBisectOrder(timestamps) {
    const arr = [];
    if (timestamps.length === 0) {
      return arr;
    } else if (timestamps.length <= 2) {
      return timestamps;
    } else {
      const beginTs = timestamps.shift();
      const endTs = timestamps.pop();
      const medianTs = timestamps.splice(timestamps.length / 2, 1)[0];
      return [beginTs, endTs, medianTs].concat(
        MrChart.sortInBisectOrder(timestamps));
    }
  }

  // Populate array of timestamps we want to fetch.
  static makeTimestamps(endDate, frequency, numDays=DEFAULT_NUM_DAYS) {
    if (!endDate) {
      throw new Error('endDate required');
    }
    const endTimeSeconds = Math.round(endDate.getTime() / 1000);
    const timestampsChronological = [];
    for (let i = 0; i < numDays; i += frequency) {
      timestampsChronological.unshift(endTimeSeconds - (SECONDS_IN_DAY * i));
    }
    return timestampsChronological;
  }

  // Convert a string '2018-11-03' to a Date object.
  static dateStringToDate(dateString) {
    if (!dateString) {
      return null;
    }
    const splitDate = dateString.split('-');
    const year = Number.parseInt(splitDate[0]);
    // Month is 0-indexed, so subtract one.
    const month = Number.parseInt(splitDate[1]) - 1;
    const day = Number.parseInt(splitDate[2]);
    return new Date(Date.UTC(year, month, day, 23, 59, 59));
  }

  // Return a URLSearchParams object. Separate method for stubbing.
  static getSearchParams() {
    // TODO(zhangtiff): Make this use page.js's queryParams object instead
    // of parsing URL params multuple times, once charts is integrated with the SPA.
    return new URLSearchParams(document.location.search.substring(1));
  }

  // Returns a Date taken from URL param, defaults to current date.
  static getEndDate() {
    const urlParams = MrChart.getSearchParams();
    if (urlParams.has('end-date')) {
      const date = MrChart.dateStringToDate(urlParams.get('end-date'));
      if (date) {
        return date;
      }
    }
    const today = new Date();
    today.setHours(23);
    today.setMinutes(59);
    today.setSeconds(59);
    return today;
  }

  // Returns a Date taken from URL param, defaults to DEFAULT_NUM_DAYS days ago.
  static getStartDate(endDate, diff) {
    const urlParams = MrChart.getSearchParams();
    if (urlParams.has('start-date')) {
      const date = MrChart.dateStringToDate(urlParams.get('start-date'));
      if (date) {
        return date;
      }
    }
    const startDate = new Date(endDate.getTime() - 1000 * SECONDS_IN_DAY * diff);
    return startDate;
  }

  static makeIndices(timestamps) {
    const dateFormat = {year: 'numeric', month: 'numeric', day: 'numeric'};
    return timestamps.map((ts) => (
      (new Date(ts * 1000)).toLocaleDateString('en-US', dateFormat)
    ));
  }

  // Generate predicted future data based on previous data.
  static getPredictedData(values, dateRange, interval, frequency) {
    // TODO(weihanl): changes to support frequencies other than 1 and 7.
    let n, endDateRange;
    if (frequency === 1) {
      // Display in daily.
      n = values.length;
      endDateRange = interval;
    } else {
      // Display in weekly.
      n = Math.floor((DEFAULT_NUM_DAYS + 1) / 7);
      endDateRange = interval * 7 - 1;
    }
    const [slope, intercept] = linearRegression(values, n);
    const endDate = new Date(MrChart.getEndDate().getTime() + 1000 * SECONDS_IN_DAY * (1 + endDateRange));
    const timestampsChronological = MrChart.makeTimestamps(endDate, frequency, endDateRange);
    const predictedIndices = MrChart.makeIndices(timestampsChronological);

    // Obtain future data and past data on the generated line.
    const predictedValues = [];
    const generatedValues = [];
    for (let i = 0; i < interval; i++) {
      predictedValues.push(Math.round(100*((i + n) * slope + intercept)) / 100);
    }
    for (let i = 0; i < n; i++) {
      generatedValues.push(Math.round(100*(i * slope + intercept)) / 100);
    }
    return [predictedIndices, predictedValues, generatedValues];
  }

  // Generate error range lines using +/- standard error on intercept to original line.
  static getErrorData(generatedValues, values, predictedValues) {
    const diffs = [];
    for (let i = 0; i < generatedValues.length; i++) {
      diffs.push(values[values.length - generatedValues.length + i] - generatedValues[i]);
    }
    const sqDiffs = diffs.map((v) => v * v);
    const stdDev = sqDiffs.reduce((sum, v) => sum + v) / values.length;
    const maxValues = predictedValues.map((x) => Math.round(100 * (x + stdDev)) / 100);
    const minValues = predictedValues.map((x) => Math.round(100 * (x - stdDev)) / 100);
    return [maxValues, minValues];
  }

  // Format all data using scattered dot representation for a single chart line.
  static getAllData(indices, values, dateRange, predRange, frequency) {
    // Set the number of data points that needs to be generated based on
    // future time range and frequency.
    let interval;
    switch(predRange) {
      case predRangeType.NEXT_MONTH:
        interval = frequency === 1 ? 30 : 4;
        break;
      case predRangeType.NEXT_QUARTER:
        interval = frequency === 1 ? 90 : 13;
        break;
      case predRangeType.NEXT_50:
        interval = Math.floor((dateRange + 1) / (frequency * 2));
        break;
    }

    const [predictedIndices, predictedValues, generatedValues] =
      MrChart.getPredictedData(values, dateRange, interval, frequency);
    const [maxValues, minValues] =
      MrChart.getErrorData(generatedValues, values, predictedValues);
    const n = generatedValues.length;

    // Format data into an array of {x:"MM/DD/YYYY", y:1.00} to draw chart.
    const originalData = [];
    const predictedData = [];
    const maxData = [{x: indices[values.length - 1], y: generatedValues[n - 1]}];
    const minData = [{x: indices[values.length - 1], y: generatedValues[n - 1]}];
    for (let i = 0; i < values.length; i++) {
      originalData.push({x:indices[i], y:values[i]});
    }
    for (let i = 0; i < n; i++) {
      predictedData.push({x:indices[values.length - n + i],
        y:Math.max(Math.round(100 * generatedValues[i]) / 100, 0)});
    }
    for (let i = 0; i < predictedValues.length; i++) {
      predictedData.push({x: predictedIndices[i], y: Math.max(predictedValues[i], 0)});
      maxData.push({x: predictedIndices[i], y: Math.max(maxValues[i], 0)});
      minData.push({x: predictedIndices[i], y: Math.max(minValues[i], 0)});
    }
    return [originalData, predictedData, maxData, minData]
  }

  // Sort lines by data in reversed chronological order
  // and return top n lines with most issues.
  static getSortedLines(arrayValues, n) {
    if (n >= arrayValues.length) {
      return arrayValues;
    }
    const len = arrayValues[0].data.length;
    // Convert data by reversing and starting from last digit and sort according to
    // the resulting value. e.g. [4,2,0] => 24, [0,4,3] => 340
    const sortedValues = arrayValues.slice().sort((arrX, arrY) => {
      const intX = parseInt(arrX.data.map((i) => i.toString()).reverse().join(''));
      const intY = parseInt(arrY.data.map((i) => i.toString()).reverse().join(''));
      return intY - intX;
    });
    return sortedValues.slice(0, n);
  }

  // Set groupby object from URL.
  static getGroupByURL() {
    const urlParams = MrChart.getSearchParams();
    if (urlParams.has('groupby')) {
      const groupBy = {value: urlParams.get('groupby')};
      switch(urlParams.get('groupby')) {
        case '':
          groupBy['display'] = 'None';
          break;
        case 'open':
          groupBy['display'] = 'Is open';
          break;
        case 'owner':
          groupBy['display'] = 'Owner';
          break;
        case 'component':
          groupBy['display'] = 'Component';
          break;
        case 'status':
          groupBy['display'] = 'Status';
          break;
        default:
          groupBy['display'] = urlParams.get('labelPrefix');
          groupBy['labelPrefix'] = urlParams.get('labelPrefix');
      }
      return groupBy;
    } else {
      return {groupBy: '', display: 'None'};
    }
  }
}

customElements.define('mr-chart', MrChart);
