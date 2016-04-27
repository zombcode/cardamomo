var Cardamomo = function() {

  this.socket = function(path) {

    var _socket
    var _actions = [];
    var _onOpen = null;
    var self = this;

    openSocket(path);

    function openSocket(path) {
      path = path.replace('http://', '');
      path = path.replace('https://', '');
      path = path.replace('ws://', '');
      path = path.replace('wss://', '');
      path = "ws://" + path;

      _socket = new WebSocket(path);
      _socket.onopen = function (event) {
        self.send("CardamomoSocketInit", "{}");

        _socket.onmessage = function (event) {
          try {
            var data = JSON.parse(event.data);
            if( data.Action == "CardamomoSocketInit" ) {
              self.id = data.Params.id;

              if(_onOpen != null) {
                _onOpen();
              }
            } else {
              for( var i in _actions ) {
                var action = _actions[i];
                if( action.action == data.Action ) {
                  action.callback(data.Params);
                }
              }
            }
          } catch(e) {}
        }

        _socket.onclose = function() {
          console.log("Disconnect!");
          //try to reconnect in 5 seconds
          setTimeout(
          function () {
            self.openSocket(this.path);
          }, 5000);
        };
      };
    }

    function onMessage(action, callback) {
      _actions.push({"action": action, "callback": callback});
    }

    function send(action, params) {
      var message = {
          "action": action,
          "params": JSON.stringify(params)
      };

      message = JSON.stringify(message);
      _socket.send(message);
    }

    this.onOpen = function(callback) {
      _onOpen = callback;
    };
    this.on = onMessage;
    this.send = send;

    return this;
  }

  return this;
}
