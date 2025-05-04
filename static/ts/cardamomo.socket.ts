export class CardamomoSocket {

  private path : string
  private ws = null;

  private id : string;
  public pingTime : Number;
  private destroyed : boolean;

  private _actions = {};
  private _onOpen = null;
  private _onClose = null;

  constructor(path : string) {
    this.pingTime = 10000;
    this.openSocket(path);
  }

  private openSocket(path) {
    path = path.replace('http://', 'ws://');
    path = path.replace('https://', 'wss://');

    if( path.indexOf('ws://') !== -1 && path.indexOf('wss://') !== -1 ) {
      path = 'ws://' + path;
    }

    this.path = path;
    this.destroyed = false;

    this.ws = new WebSocket(this.path);
    this.ws.onopen = (event) => {
      this.send("CardamomoSocketInit", "{}");
    };

    this.ws.onmessage = (event) => {
      try {
        var data = JSON.parse(event.data);
        if( data.Action == "CardamomoSocketInit" ) {
          this.id = data.Params.id;

          if(this._onOpen != null) {
            this._onOpen();
          }

          this.ping();
        } else if( data.Action == "CardamomoPong" ) {
          // Pong
        } else {
          if( data.Action in this._actions ) {
            this._actions[data.Action](data.Params);
          }
        }
      } catch(e) {}
    }

    this.ws.onclose = () => {
      if(this._onClose != null) {
        this._onClose();
      }
      //try to reconnect in 5 seconds
      if( this.destroyed == false ) {
        setTimeout(
        () => {
          this.openSocket(this.path);
        },5000);
      }
    };
  }

  send(action, params) {
    var message = {
        "action": action,
        "params": JSON.stringify(params)
    };

    var messageStr = JSON.stringify(message);
    this.ws.send(messageStr);
  }

  ping() {
    if( this.destroyed == false ) {
      this.send("CardamomoPing", "{}");

      setTimeout(() => {
        this.ping();
      }, this.pingTime);
    }
  }

  destroy() {
    this.destroyed = true;
    this.ws.close();
  }

  on = (action, callback) => {
    this._actions[action] = callback;
  }

  onOpen = (callback) => {
    this._onOpen = callback;
  };

  onClose = (callback) => {
    this._onClose = callback;
  };
}
