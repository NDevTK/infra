function load_prefs () {
  console.log("load_prefs");
  chrome.storage.local.get("tracking_branch", function(tb) {
    var tb_input = document.getElementById("tracking_branch");
    tb_input.focus();
    tb_input.addEventListener("change", set_tracking_branch);
    if (tb && tb_input && tb["tracking_branch"]) {
      tb_input.value = tb["tracking_branch"];
    }
  });
}

function set_tracking_branch () {
  console.log("set_tracking_branch");
  var tb_input = document.getElementById("tracking_branch");
  var tb = tb_input.value;
  console.log("tb=" + tb);
  chrome.storage.local.set({"tracking_branch": tb});
}

document.addEventListener("DOMContentLoaded", load_prefs);