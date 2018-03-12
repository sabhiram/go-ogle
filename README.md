# go-ogle

A command line way to search google in a chrome tab.

## What?

A minimal command line app + a chrome extension which allows you to:
1. Search Google for something
2. Navigate using [^ v]
3. Select a link using [Enter]

## Install

TODO

## Contents

* `extension` implements the chrome extension implemented in Javascript.
* `hub` implements a golang library that implements a pub-sub socket.
* `server` implements a golang websocket server library.
* `types` contain application specific types that are usually passed around library instances.
* `main.go` implements the cli app which will self-spawn a daemon process to connect to.

## TODOs

1. Advanced keyboard input cases - next page / prev page / next-result from page etc
2. If selected item is not focused on page - center it.
