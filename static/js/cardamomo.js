var Cardamomo = function() {

  this.socket = function(path) {

    var id = "";
    var destroyed = false;

    var _socket = null;
    var _actions = {};
    var _onOpen = null;
    var _onClose = null;
    var self = this;

    self.openSocket = function (path) {
      path = path.replace('http://', '');
      path = path.replace('https://', '');
      path = path.replace('ws://', '');
      path = path.replace('wss://', '');
      path = "ws://" + path;

      self.destroyed = false;
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

            self.ping();
          } else if( data.Action == "CardamomoPong" ) {
            // Pong
          } else {
            if( data.Action in _actions ) {
              _actions[data.Action](data.Params);
            }
          }
        } catch(e) {}
      }

      _socket.onclose = (function() {
      	if(_onClose != null) {
		  _onClose();
		}
        //try to reconnect in 5 seconds
        if( self.destroyed == false ) {
          setTimeout((function () {
            self._socket = null;
            self.openSocket(path);
          }).bind(path), 5000);
        }
      }).bind(path);
    }

    function onMessage(action, callback) {
      _actions[action] = callback;
    }

    function send(action, params) {
      var message = {
          "action": action,
          "params": JSON.stringify(params)
      };

      message = JSON.stringify(message);
      _socket.send(message);
    }

    function ping() {
      if( self.destroyed == false ) {
        self.send("CardamomoPing", "{}");

        setTimeout(function () {
          self.ping();
        }, self.pingTime);
      }
    }

    function destroy() {
      self.destroyed = true;
      self._socket.close();
    }

    this.onOpen = function(callback) {
      _onOpen = callback;
    };
    this.onClose = function(callback) {
      _onClose = callback;
    };
    this.on = onMessage;
    this.send = send;
    this.ping = ping;
    this.pingTime = 10000;

    this.openSocket(path);

    return this;
  }

  return this;
}
