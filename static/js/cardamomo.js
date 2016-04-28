var Cardamomo = function() {

  this.socket = function(path) {

    var _socket = null;
    var _actions = [];
    var _onOpen = null;
    var self = this;

    self.openSocket = function (path) {
      path = path.replace('http://', '');
      path = path.replace('https://', '');
      path = path.replace('ws://', '');
      path = path.replace('wss://', '');
      path = "ws://" + path;

      _socket = new WebSocket(path);

      _socket.onopen = (function (event) {
        self.send("CardamomoSocketInit", "{}");
      });

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

      _socket.onclose = (function() {
        console.log("Disconnect!");
        //try to reconnect in 5 seconds
        setTimeout((function () {
          self._socket = null;
          self.openSocket(path);
        }).bind(path), 5000);
      }).bind(path);
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

    this.openSocket(path);

    return this;
  }

  return this;
}
