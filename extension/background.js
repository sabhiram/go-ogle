////////////////////////////////////////////////////////////////////////////////

let last_port = null
  , ws 		  = null
  , last_win  = -1
  , last_tab  = -1
  ;

////////////////////////////////////////////////////////////////////////////////

// Called when the user clicks on the browser action.
chrome.browserAction.onClicked.addListener(function(tab) {

	// TODO: Popup / info window when app-icon is clicked.

});

////////////////////////////////////////////////////////////////////////////////

getSingleton = function(func) {
    return function() {
		var base, pf, ref;
		let args = (1 <= arguments.length) ? slice.call(arguments, 0) : [];
		ref = [func, null], pf = ref[0], func = ref[1];
		base = ref;
		return typeof base[0] === "function" ? base[0].apply(base, args) : void 0;
    };
}

tryReconnect = function() {
	last_port = null;
	last_tab  = -1;
	last_win  = -1;
	console.log("Try again in 1000ms");
	setTimeout(connectToServer, 1000);
}

portSend = function(obj) {
	if (last_port) {
		last_port.postMessage(obj);
	}
}

// Function that connects to the `go-ogle` server and listens for commands
// to run in browser.
connectToServer = function() {
	let reconnect = getSingleton(tryReconnect);

	const url = "ws://localhost:18881/ws";
	
	try {
		ws = new WebSocket(url);
	} catch(e) {
	}


	ws.onopen = function() {
		// socket is open ...
	};
	ws.onerror = function(err) {
		reconnect();
	};
	ws.onclose = function(err) {
		reconnect();
	};
	ws.onmessage = function(e) {
		let data = JSON.parse(e.data);
		if (data["Type"] == "open_new_tab_with_url") {
			chrome.tabs.create({ url: data["Data"] }, function(tab) {
				last_tab = tab.id;
				last_win = tab.windowId;
				chrome.tabs.executeScript(tab.id, {file: "content_script.js"});
			});
		}
		else if (data["Type"] == "select_current_result") {
			if (last_port) {
				last_port.postMessage({command: "select_current_result"});
				if (last_win > 0 && last_tab > 0) {
					chrome.windows.update(last_win, {"focused": true, "drawAttention": true});
				}
			}
		}
		else {
			// All other message types are blindly forwarded to the downstream page 
			// of search results (the last one that connected).
			if (last_port) {
				last_port.postMessage({command: data["Type"], data: data["Data"]})
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













