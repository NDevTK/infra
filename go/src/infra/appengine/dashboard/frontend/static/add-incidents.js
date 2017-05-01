// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
(function(window) {
  'use strict';

  const Alert = {
    RED: 0,
    YELLOW: 1,
  };

  const IconType = {
    CIRCLE: "circle",
    LEFT: "left",
    RIGHT: "right",
    CENTER: "center",
  };

  let alertImgsMap = {};
  alertImgsMap[Alert.RED] = {
    "circle": 'static/red.png',
    "left": 'static/red_left.png',
    "right": 'static/red_right.png',
    "center": 'static/red_rect.png',
  };
  alertImgsMap[Alert.YELLOW] = {
    "center": 'static/yellow.png',
    "left": 'static/yellow_left.png',
    "right": 'static/yellow_right.png',
    "center": 'static/yellow_rect.png',
  };

  function addIncidents(pageData) {
    renderIncidents(
	pageData['ChopsServices'],
	pageData['Dates'][0],
	pageData['Dates'][6]);
    renderIncidents(
	pageData['NonSLAServices'],
	pageData['Dates'][0],
	pageData['Dates'][6]);
  }

  function renderIncidents(services, firstDate, lastDate) {
    for (let i = 0; i < services.length; i++) {
      let service = services[i];
      let serviceName = service['Service']['Name'];
      for (let j = 0; j < service['Incidents'].length; j++) {
	let incident = service['Incidents'][j];
	let img = getIncidentImg(incident['Severity'], IconType.CIRCLE);
	if (incident['Open']) {
	  let statusCell = document.querySelector('.js-' + serviceName);
	  statusCell.appendChild(getIncidentImg(
	      incident['Severity'], IconType.CIRCLE));
	} else {
	  addIncident(incident, serviceName, firstDate, lastDate);
	}
      }
    }
  }

  function addIncident(incident, serviceName, firstDate, lastDate) {
    var startDateCell = getDateCell(incident['StartTime'], serviceName);
    if (!startDateCell) {
      startDateCell = getDateCell(firstDate, serviceName);
      startDateCell.appendChild(
	  getIncidentImg(incident['Severity'], IconType.CENTER));
    } else {
      startDateCell.appendChild(
	  getIncidentImg(incident['Severity'], IconType.LEFT));
    }
    var endDateCell = getDateCell(incident['EndTime'], serviceName);
    if (!endDateCell) {
      endDateCell = getDateCell(lastDate, serviceName);
      endDateCell.appendChild(
	  getIncidentImg(incident['Severity'], IconType.CENTER));
    } else {
      endDateCell.appendChild(
	  getIncidentImg(incident['Severity'], IconType.RIGHT));
    }
    //TODO(jojwang): add red_rect/yellow_rect to fill in
    //space between startDateCell and endDateCell
  }

  function getDateCell(date, serviceName) {
    let className = '.js-' + serviceName + '-' + fmtDate(date);
    return document.querySelector(className);
  }

  function getIncidentImg(severity, iconType) {
    let img = document.createElement('img');
    img.classList.add('light');
    img.src = alertImgsMap[severity][iconType];
    return img;
  }

  function fmtDate(rawDate) {
    let date = new Date(rawDate);
    return (date.getMonth()+1) + '-' +
	date.getDate() + '-' + date.getFullYear();
  }

  window.__addIncidents = window.__addIncidents || addIncidents;
})(window);
