
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
		// TODO: Retry
	};
	ws.onclose = function(err) {
		console.log("Socket error: ", err);
		// TODO: Retry	
	};
	ws.onmessage = function(e) {
		console.log(e);
		let data = JSON.parse(e.data);
		console.log(data)
		if (data["Type"] == "CHROME_COMMAND") {
			chrome.tabs.create({ url: data["Data"] });
		}
	};
}

////////////////////////////////////////////////////////////////////////////////

connectToServer();

////////////////////////////////////////////////////////////////////////////////
