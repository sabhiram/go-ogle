/*
 *  This content is injected by the `go-ogle` chrome extension for the 
 *  following purposes:
 *	
 *  1.  Open a socket to communicate with the extension.
 *  2.  React to requests from the extension to select the next / prev
 *	    search result (and apply appropriate styling).
 *  3.  This should also only be applied to pages that match the 
 * 		google search page URL prefix.
 */

var port = chrome.runtime.connect({name: "go-ogle"})
  , indx = 0
  , curr = null
  ;

function selectResult(el) {
	el.style.backgroundColor = "#3232d240";
}

function unselectResult(el) {
	el.style.backgroundColor = "transparent";
}

function setSelectedResult(idx) {
	let results = document.querySelectorAll("#search .g");
	if (results.length > idx) {
		if (curr != null) {
			unselectResult(curr);
		}

		indx = idx;
		curr = results[indx];
		selectResult(curr);
	}
}

port.onMessage.addListener(function(msg) {
	if (!msg.command) return;

	switch (msg.command) {
	case "select_result":
		if (msg.slot != undefined && typeof msg.slot == "number") {
			setSelectedResult(msg.slot);
		}
		break;
	case "select_next_result":
		setSelectedResult(indx + 1);
		break;
	case "select_prev_result":
		setSelectedResult(indx - 1);
		break;
	default:
		port.postMessage({type: "error", msg: "invalid command specified"});
		break;
	}
});
