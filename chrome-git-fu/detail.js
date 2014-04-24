// Copyright (c) 2014 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.


(function () {
  "use strict";
  var sha1_re = /[a-f0-9]{40}/i;
  var commit_url_re = new RegExp(
      "^(https://.*\.googlesource\.com/[^+]*)/[+]/[a-f0-9]{40}$", "i");
  var json_prefix_re = new RegExp("^\)\]}'");

  function gitiles_log_load () {
    if (this.status != 200) {
      return;
    }
    var json_str = this.responseText.replace(json_prefix_re, "");
    var j = JSON.parse(json_str);
    var cls = "contains_loading";
    if (j && j["log"].length > 0) {
      cls = "contains_no";
    } else if (j && j["log"].length == 0) {
      cls = "contains_yes";
    }
    this.target_span.setAttribute("class", cls);
  }

  function gitiles_log_error () {
    console.log(this.status);
    console.log(this.statusText);
  }

  function branch_contains(gitiles_url, sha1, branch) {
    var new_span = document.createElement('span');
    new_span.setAttribute('class', 'contains_loading');
    var new_text = document.createTextNode(" (" + branch + ")");
    new_span.appendChild(new_text);
    var log_url = gitiles_url + "/+log/" + branch + ".." + sha1
        + "?format=JSON&n=1";
    var req = new XMLHttpRequest();
    req.target_span = new_span;
    req.onload = gitiles_log_load;
    req.onerror = gitiles_log_error;
    req.open("get", log_url, true);
    req.send();
    return new_span;
  }

  function annotate_shas(node, tb) {
    var new_nodes = new Array();
    var modified = false;
    var gitiles_url = null;

    // First pass: look for gitiles URL
    for (var j = 0; j < node.childNodes.length; j++) {
      var n = node.childNodes[j];
      if (n.nodeType == Node.ELEMENT_NODE && n.tagName == "A") {
        var match = (n.getAttribute("href") || "").match(commit_url_re);
        if (match) {
          gitiles_url = match[1];
          break;
        }
      }
    }

    // Second pass: Find/annotate sha1's.
    for (var j = 0; j < node.childNodes.length; j++) {
      var n = node.childNodes[j];
      // Discard previous annotations.
      if (n.nodeType == Node.ELEMENT_NODE && n.tagName == "SPAN" &&
          (n.getAttribute("class") || "").substr(0, 9) == "contains_") {
        continue;
      }
      if (n.nodeType != Node.TEXT_NODE) {
        new_nodes.push(n);
        continue;
      }
      var start_pos = 0;
      var match_pos = n.wholeText.search(sha1_re);
      while (match_pos != -1) {
        var sha1_str = n.wholeText.substr(match_pos, 40);
        modified = true;
        var new_text = document.createTextNode(
            n.wholeText.substr(start_pos, match_pos - start_pos + 40));
        new_nodes.push(new_text);
        if (gitiles_url) {
          // If this is a re-annotation after tracking_branch changed, don't
          // recompute information for "master".
          if (match_pos + 40 == n.wholeText.length &&
              n.nextSibling && n.nextSibling.nodeType == Node.ELEMENT_NODE &&
              n.nextSibling.tagName == "SPAN" &&
              (n.nextSibling.getAttribute("class") || "").substr(0, 9) == "contains_" &&
              n.nextSibling.firstChild.wholeText == " (master)") {
            new_nodes.push(n.nextSibling);
          } else {
            new_nodes.push(branch_contains(gitiles_url, sha1_str, "master"));
          }
          if (tb) {
            new_nodes.push(
                branch_contains(gitiles_url, sha1_str, "branch-heads/" + tb));
          }
        }
        start_pos = match_pos + 40;
        match_pos = n.wholeText.substr(start_pos).search(sha1_re);
      }

      if (start_pos == 0) {
        new_nodes.push(n);
      } else if (start_pos < n.wholeText.length) {
        new_text = document.createTextNode(
            n.wholeText.substr(start_pos));
        new_nodes.push(new_text);
      }
    }
    if (!modified) {
      return;
    }
    while (node.firstChild) {
      node.removeChild(node.firstChild);
    }
    for (var j = 0; j < new_nodes.length; j++) {
      node.appendChild(new_nodes[j]);
    }
  }

  chrome.runtime.sendMessage(null, "show");

  function annotate_page () {
    chrome.storage.local.get("tracking_branch", function(tb) {
      var tracking_branch = tb["tracking_branch"];
      var comments = document.getElementsByClassName('cursor_off vt issuecomment');
      for (var i = 0; i < comments.length; i++) {
        var el = comments.item(i).getElementsByTagName('pre');
        if (el.length != 1) {
          continue;
        }
        annotate_shas(el[0], tracking_branch);
      }
    });
  }

  chrome.runtime.onMessage.addListener(function (message, sender, sendResponse) {
    if (message == 'tracking_branch_changed') {
      setTimeout(function() { annotate_page() }, 0);
    }
  });

  annotate_page();
})();
