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

  var lapCount = document.getElementById('lapCount')
  var currentLap = document.getElementById('currentLapTime');
  var previousLap = document.getElementById('previousLapTime');
  var bestLap = document.getElementById('bestLapTime');
  var pbLap = document.getElementById('pbLapTime');

  var checkEngineLightAlert = document.getElementById('celAlert');
  var dataLoggingAlert = document.getElementById('dataloggingAlert');

  socket.onmessage = function(event) {
    const data = JSON.parse(event.data);

    switch (data.Type) {
      case 1:
        switch (data.FrameId) {
          case 660:
            rpmBar.style.width = ((data.Rpm / 9000) * 100) + '%';
            rpmNum.textContent = data.Rpm;
            speed.textContent = data.Speed;
            gear.textContent = data.Gear;
            voltage.textContent = data.Voltage;
            break;
          case 661:
            iat.textContent = data.Iat;
            ect.textContent = data.Ect;
            break;
          case 662:
            tpsBar.style.height = data.Tps + '%';
            break;
          case 664:
            lambdaRatio.textContent = data.LambdaRatio;
            break;
          case 667:
            oilTemp.textContent = data.OilTemp;
            oilPressure.textContent = data.OilPressure;
            break;
        }
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

        lapCount.textContent = data.LapCount;
        previousLap.textContent = `${previousLapMinutes}:${previousLapSeconds}.${previousLapMilliseconds}`;
        bestLap.textContent = `${bestLapMinutes}:${bestLapSeconds}.${bestLapMilliseconds}`;
        pbLap.textContent = `${pbLapMinutes}:${pbLapSeconds}.${pbLapMilliseconds}`;
        break;
      
      case 4:
        if (data.AlertCoolantTemp)
          ect.style.setProperty('background-color', 'red')
        else
          ect.style.setProperty('background-color', 'transparent')

        if (data.AlertOilTemp)
          oilTemp.style.setProperty('background-color', 'red')
        else
          oilTemp.style.setProperty('background-color', 'transparent')

        if (data.AlertOilPressure)
          oilPressure.style.setProperty('background-color', 'red')
        else
          oilPressure.style.setProperty('background-color', 'transparent')
        break;
      
      case 5:
        if (data.CELAlert) {
          checkEngineLightAlert.style.setProperty('background-color', 'red');
          checkEngineLightAlert.classList.add('cel-blinker') 
        } else {
          checkEngineLightAlert.style.animation = "";
        }
        
        if (data.DataloggingAlert)
          dataLoggingAlert.style.setProperty('background-color', 'teal');
        else
          dataLoggingAlert.style.setProperty('background-color', 'transparent')

        if (data.ChangePage)
          window.location.href = 'http://localhost:8080/ShiftLightDisplay/'
        
        break;
    }
  }
});