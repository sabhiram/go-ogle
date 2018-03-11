////////////////////////////////////////////////////////////////////////////////

let last_port = null
  ;

////////////////////////////////////////////////////////////////////////////////

// Called when the user clicks on the browser action.
chrome.browserAction.onClicked.addListener(function(tab) {

	// TODO: Popup / info window when app-icon is clicked.

});

////////////////////////////////////////////////////////////////////////////////

// Function that connects to the `go-ogle` server and listens for commands
// to run in browser.
connectToServer = function() {
	const url = "ws://localhost:18881/ws";
	ws = new WebSocket(url);

	ws.onopen = function() {
		console.log("Socket open");
	};
	ws.onerror = function(err) {
		console.log("Socket error: ", err);
		last_port = null;
		// TODO: Retry
	};
	ws.onclose = function(err) {
		console.log("Socket error: ", err);
		last_port = null;
		// TODO: Retry	
	};
	ws.onmessage = function(e) {
		let data = JSON.parse(e.data);
		if (data["Type"] == "open_new_tab_with_url") {
			chrome.tabs.create({ url: data["Data"] }, function(tab) {
				chrome.tabs.executeScript(tab.id, {file: "content_script.js"});
			});
		}
		else if (data["Type"] == "next_result") {
			if (last_port) {
				last_port.postMessage({command: "highlight_next_result"});
			}
		}
		else if (data["Type"] == "prev_result") {
			if (last_port) {
				last_port.postMessage({command: "highlight_prev_result"});
			}
		}
		else if (data["Type"] == "select_current_result") {
			if (last_port) {
				last_port.postMessage({command: "select_current_result"});
			}
		}
	};
}

////////////////////////////////////////////////////////////////////////////////

// Connect to the websocket server that will send us commands.
connectToServer();

// Setup a listener so we can catch chrome sockets connecting to the extension
// from a page which has `content_script.js` injected.
chrome.runtime.onConnect.addListener(function(port) {
	console.assert(port.name == "go-ogle");

	port.onMessage.addListener(function(msg) {
		console.log("Page => Extension : ", msg);
	});

	last_port = port;
	last_port.postMessage({command: "highlight_result", slot: 0});
});

////////////////////////////////////////////////////////////////////////////////
