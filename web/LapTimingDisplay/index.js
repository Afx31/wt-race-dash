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
  var gear = document.getElementById('gear');
  var voltage = document.getElementById('voltage');
  var iat = document.getElementById('iat');
  var ect = document.getElementById('ect');
  var tpsBar = document.getElementById('tpsbar');
  var lambdaRatio = document.getElementById('lambdaRatio');
  var oilTemp = document.getElementById('oilTemp');
  var oilPressure = document.getElementById('oilPressure');

  var currentLap = document.getElementById('currentLapTime');
  var previousLap = document.getElementById('previousLapTime');
  var bestLap = document.getElementById('bestLapTime');
  var pbLap = document.getElementById('pbLapTime');

  socket.onmessage = function(event) {
    const data = JSON.parse(event.data);
    
    switch (data.Type) {
      case 1:
        // RPM progressive bar
        rpmBar.style.setProperty('max-width', '1920px', 'important');
        var rpmbarPercentage = (data.Rpm / 9000) * 100;

        tpsBar.style.height = data.Tps + '%';
        rpmBar.style.width = `${rpmbarPercentage}%`;
        rpmNum.textContent = data.Rpm;
        speed.textContent = data.Speed;
        gear.textContent = data.Gear;
        voltage.textContent = (data.Voltage / 10).toFixed(1);
        iat.textContent = data.Iat;
        ect.textContent = data.Ect;
        lambdaRatio.textContent = data.LambdaRatio;
        oilTemp.textContent = data.OilTemp;
        oilPressure.textContent = data.OilPressure;
        break;
      case 2:
        var currentLapMinutes = Math.floor((data.CurrentLapTime % 3600000) / 60000);
        var currentLapSeconds = (Math.floor((data.CurrentLapTime % 60000) / 1000)).toString().padStart(2, '0');
        var currentLapMilliseconds = (data.CurrentLapTime % 1000).toString().padStart(3, '0');
        
        currentLap.textContent = `${currentLapMinutes}:${currentLapSeconds}.${currentLapMilliseconds}`;
        break;
      case 3:
        var previousLapMinutes = Math.floor((data.PreviousLapTime % 3600000) / 60000);
        var previousLapSeconds = (Math.floor((data.PreviousLapTime % 60000) / 1000)).toString().padStart(2, '0');
        var previousLapMilliseconds = (data.PreviousLapTime % 1000).toString().padStart(3, '0');

        var bestLapMinutes = Math.floor((data.BestLapTime % 3600000) / 60000);
        var bestLapSeconds = (Math.floor((data.BestLapTime % 60000) / 1000)).toString().padStart(2, '0');
        var bestLapMilliseconds = (data.BestLapTime % 1000).toString().padStart(3, '0');

        var pbLapMinutes = Math.floor((data.PbLapTime % 3600000) / 60000);
        var pbLapSeconds = (Math.floor((data.PbLapTime % 60000) / 1000)).toString().padStart(2, '0');
        var pbLapMilliseconds = (data.PbLapTime % 1000).toString().padStart(3, '0');

        previousLap.textContent = `${previousLapMinutes}:${previousLapSeconds}.${previousLapMilliseconds}`;
        bestLap.textContent = `${bestLapMinutes}:${bestLapSeconds}.${bestLapMilliseconds}`;
        pbLap.textContent = `${pbLapMinutes}:${pbLapSeconds}.${pbLapMilliseconds}`;
        break;
    }
  }
});