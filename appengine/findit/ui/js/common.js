/* Copyright 2017 The Chromium Authors. All Rights Reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file or at
 * https://developers.google.com/open-source/licenses/bsd
 */

/**
 * Shows the given message via <app-messages>.
 *
 * @param {number} messageId The ID of a predefined message in <app-messages>.
 * @param {string} content The content of the customized message.
 * @param {string} titile The title for the predefined or customized message.
 * @param {boolean} preFormat If true, show the message in a <pre> tag.
 */
function displayMessage(messageId, content, title, preFormat) {
  var detail = {
    'messageId': messageId,
    'content': content,
    'title': title,
    'preFormat': preFormat,
  };
  var event = new CustomEvent('message', {'detail': detail});
  console.log('Dispatching message event:');
  console.log(event);
  document.dispatchEvent(event);
}

function shortenTimeDelta(long_time_delta) {
  var pattern = /(\d day[s]?)?,?\s?(\d*):(\d*):(\d*)/;
  // [full match, n day(s), HH, MM, SS]
  var res = long_time_delta.match(pattern);

  if (typeof(res[1]) != 'undefined') {
    return res[1];
  }

  var index_to_rep = {
    2: 'hour',
    3: 'minute',
    4: 'second'
  };

  for (var i=2; i<res.length; i++) {
      var int_rep = parseInt(res[i]);
      if ( int_rep == 1) {
        return int_rep + ' ' + index_to_rep[i];
      }
      if ( int_rep > 1) {
        return int_rep + ' ' + index_to_rep[i] + 's';
      }
  }
  return 'just now';
}
