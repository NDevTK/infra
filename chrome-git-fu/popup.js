var tracking_branch = null;

function load_prefs () {
  console.log("load_prefs");
  chrome.storage.local.get("tracking_branch", function(tb) {
    tracking_branch = tb["tracking_branch"];
    var tb_input = document.getElementById("tracking_branch");
    if (tb && tb_input && tracking_branch) {
      tb_input.value = tracking_branch;
    }
    tb_input.focus();
    tb_input.addEventListener("change", set_tracking_branch);
  });
}

function set_tracking_branch () {
  function notify_tabs () {
    query_params = {'active': true, 'currentWindow': true};
    chrome.tabs.query(query_params, function (tab_array) {
      for (var i = 0; i < tab_array.length; i++) {
	chrome.tabs.sendMessage(tab_array[i].id, 'tracking_branch_changed');
      }
    });
  }
  console.log("set_tracking_branch");
  var tb_input = document.getElementById("tracking_branch");
  if (tracking_branch != tb_input.value) {
    tracking_branch = tb_input.value;
    chrome.storage.local.set(
	{"tracking_branch": tracking_branch}, notify_tabs);
  }
}

document.addEventListener("DOMContentLoaded", load_prefs);