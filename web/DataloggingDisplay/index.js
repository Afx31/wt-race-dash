const socket = new WebSocket('ws://' + document.location.host + '/ws');

socket.onopen = function(event) {
  console.log('Connected to WebSocket server.');
};
socket.onclose = function (event) {
  console.log('Connection closed: ', event);
};
socket.onerror = function(err) {
  console.log('WebSocket error:', err);
};
window.addEventListener("beforeunload", function () {
  socket.close();
});

document.addEventListener('DOMContentLoaded', () => {
  var rpmBar = document.getElementById('rpmbar');
  var rpmNum = document.getElementById('rpmNum');
  var speed = document.getElementById('speed');
  var voltage = document.getElementById('voltage');
  var iat = document.getElementById('iat');
  var ect = document.getElementById('ect');
  var tps = document.getElementById('tps');
  var map = document.getElementById('map');
  var inj = document.getElementById('inj');
  var ign = document.getElementById('ign');
  var lambdaRatio = document.getElementById('lambdaRatio');
  var oilTemp = document.getElementById('oilTemp');
  var oilPressure = document.getElementById('oilPressure');

  socket.onmessage = function(event) {
    const data = JSON.parse(event.data);

    switch (data.Type) {
      case 1:
        rpmBar.style.width = ((data.Rpm / 9000) * 100) + '%';
        rpmNum.textContent = data.Rpm;
        speed.textContent = data.Speed;
        voltage.textContent = data.Voltage;
        iat.textContent = data.Iat;
        ect.textContent = data.Ect;
        tps.textContent = data.Tps;
        map.textContent = data.Map;
        inj.textContent = data.Inj;
        ign.textContent = data.Ign;
        lambdaRatio.textContent = data.LambdaRatio;
        oilTemp.textContent = data.OilTemp;
        oilPressure.textContent = data.OilPressure;
        
        break;
      case 5:
        if (data.ChangePage)
          window.location.href = 'http://localhost:8080/LapTimingDisplay/'
        break;
    }
  };
});