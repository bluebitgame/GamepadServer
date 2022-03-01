const configs = {
	// Chrome
	//'Xbox 360 Controller (XInput STANDARD GAMEPAD)': [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16],
	'Xbox Wireless Controller (STANDARD GAMEPAD Vendor: 045e Product: 02fd)': [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16],
    'Pro Controller (STANDARD GAMEPAD Vendor: 057e Product: 2009)': [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17],
    'Wireless Gamepad (STANDARD GAMEPAD Vendor: 057e Product: 2009)': [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17],
    
    // Safari
    'Xbox Wireless Controller Extended Gamepad': [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17],
    'Pro Controller Extended Gamepad': [1, 0, 3, 2, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17],
}

var haveEvents = 'GamepadEvent' in window;
var haveWebkitEvents = 'WebKitGamepadEvent' in window;
var controllers = {};
var sockets = {};
var rAF = window.mozRequestAnimationFrame ||
  window.webkitRequestAnimationFrame ||
  window.requestAnimationFrame;

function connecthandler(e) {
  addgamepad(e.gamepad);
}
function addgamepad(gamepad) {
  if (!configs.hasOwnProperty(gamepad.id)) {
    return;
  }
  
  var location = window.location;
  controllers[gamepad.index] = gamepad;
  var websocketurl = "ws://" + location.hostname + ":" + location.port + "/ws";
  var socket = new WebSocket(websocketurl);
  socket.addEventListener('open', function (event) {
    socket.send(gamepad.id);
    sockets[gamepad.index] = socket;
  });

  var d = document.createElement("div");
  d.setAttribute("id", "controller" + gamepad.index);
  var t = document.createElement("h1");
  t.appendChild(document.createTextNode(gamepad.id));
  d.appendChild(t);
  var b = document.createElement("div");
  b.className = "buttons";
  for (var i=0; i<gamepad.buttons.length; i++) {
    var e = document.createElement("span");
    e.className = "button";
    //e.id = "b" + i;
    e.innerHTML = i;
    b.appendChild(e);
  }
  d.appendChild(b);
  var a = document.createElement("div");
  a.className = "axes";
  for (i=0; i<gamepad.axes.length; i++) {
    e = document.createElement("meter");
    e.className = "axis";
    //e.id = "a" + i;
    e.setAttribute("min", "-1");
    e.setAttribute("max", "1");
    e.setAttribute("value", "0");
    e.innerHTML = i;
    a.appendChild(e);
  }
  d.appendChild(a);
  document.getElementById("start").style.display = "none";
  document.body.appendChild(d);
  rAF(updateStatus);
}

function disconnecthandler(e) {
  removegamepad(e.gamepad);
}

function removegamepad(gamepad) {
  var d = document.getElementById("controller" + gamepad.index);
  document.body.removeChild(d);
  delete controllers[gamepad.index];
  sockets[gamepad.index].close();
  delete sockets[gamepad.index];
}

function updateStatus() {
  scangamepads();
  for (j in controllers) {
    var controller = controllers[j];
	var mapping = configs[controller.id]
    var msg = {}
    var d = document.getElementById("controller" + j);
	
	var buttons_value = new Array();
    var buttons = d.getElementsByClassName("button");
    for (var i=0; i<controller.buttons.length; i++) {
      var b = buttons[i];
      var val = controller.buttons[i];
      var pressed = val == 1.0;
      var touched = false;
      if (typeof(val) == "object") {
        pressed = val.pressed;
        if ('touched' in val) {
          touched = val.touched;
        }
        val = val.value;
      }
      var pct = Math.round(val * 100) + "%";
      b.style.backgroundSize = pct + " " + pct;
      b.className = "button";
      if (pressed) {
        b.className += " pressed";
      }
      if (touched) {
        b.className += " touched";
      }
	  
      buttons_value[mapping[i]] = Math.floor(val * 255);
    }
    msg['buttons'] = buttons_value;

    var axes_value = new Array();
    var axes = d.getElementsByClassName("axis");
    for (var i=0; i<controller.axes.length; i++) {
      var a = axes[i];
      a.innerHTML = i + ": " + controller.axes[i].toFixed(4);
      a.setAttribute("value", controller.axes[i]);
    
	  axes_value[i] = Math.floor(controller.axes[i] * 32767);
    }
    msg['axes'] = axes_value;
    
    var socket = sockets[j];
    if (socket) {
      socket.send(JSON.stringify(msg));
    }
  }

  rAF(updateStatus);
}

function scangamepads() {
  var gamepads = navigator.getGamepads ? navigator.getGamepads() : (navigator.webkitGetGamepads ? navigator.webkitGetGamepads() : []);
  for (var i = 0; i < gamepads.length; i++) {
    if (gamepads[i] && (gamepads[i].index in controllers)) {
      controllers[gamepads[i].index] = gamepads[i];
    }
  }
}

if (haveEvents) {
  window.addEventListener("gamepadconnected", connecthandler);
  window.addEventListener("gamepaddisconnected", disconnecthandler);
} else if (haveWebkitEvents) {
  window.addEventListener("webkitgamepadconnected", connecthandler);
  window.addEventListener("webkitgamepaddisconnected", disconnecthandler);
} else {
  setInterval(scangamepads, 500);
}
