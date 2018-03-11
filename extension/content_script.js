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

function getResult(idx) {
	let results = document.querySelectorAll("#search .srg .g .rc");
	if (results.length > idx) return results[idx];
	return null;
}

function highlighResult(el) {
	el.style.backgroundColor = "#3232d240";
}

function unhighlighResult(el) {
	el.style.backgroundColor = "transparent";
}

function setHighlightedResult(idx) {
	let el = getResult(idx);
	if (!el) return;

	if (curr != null) {
		unhighlighResult(curr);
	}

	indx = idx;
	curr = el;
	highlighResult(curr);
}

function setPageURL(idx) {
	let el = getResult(idx);
	if (!el) return;

	let a = el.querySelector("h3.r a");
	
	window.location = a.href;
}

port.onMessage.addListener(function(msg) {
	console.log(msg);
	if (!msg.command) return;

	switch (msg.command) {
	case "highlight_result":
		if (msg.slot != undefined && typeof msg.slot == "number") {
			setHighlightedResult(msg.slot);
		}
		break;
	case "highlight_next_result":
		setHighlightedResult(indx + 1);
		break;
	case "highlight_prev_result":
		setHighlightedResult(indx - 1);
		break;
	case "select_current_result":
		setPageURL(indx);
		break;
	case "clear_selected":
		unhighlighResult(curr);
		indx = 0;
		curr = null;
		break;
	default:
		port.postMessage({type: "error", msg: "invalid command specified"});
		break;
	}
});
