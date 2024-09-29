var socket = io.connect('localhost:3000');

socket.on('CANBusMessage', (data) => {
  // if (data.changeDisplay === 1)
  //   window.location.href = 'http://localhost:3000/DataLoggingDisplay';
  
  var rpmBar = document.getElementById('rpmbar');
  var rpmNum = document.getElementById('rpmNum');
  var speed = document.getElementById('speed');
  var voltage = document.getElementById('voltage');
  var iat = document.getElementById('iat');
  var ect = document.getElementById('ect');
  var tps = document.getElementById('tps');
  var oilTemp = document.getElementById('oilTemp');
  var oilPressure = document.getElementById('oilPressure');

  // RPM progressive bar
  rpmBar.style.setProperty('max-width', '1920px', 'important');
  var rpmbarPercentage = (data.rpm / 9000) * 100;

  // Assign data to UI controls
  rpmBar.style.width = `${rpmbarPercentage}%`;
  rpmNum.textContent = data.rpm;
  speed.textContent = data.speed;
  voltage.textContent = (data.voltage / 10).toFixed(1);
  iat.textContent = data.iat;
  ect.textContent = data.ect;
  tps.textContent = data.tps;
  oilTemp.textContent = data.oilTemp;
  oilPressure.textContent = data.oilPressure;

  // RPM Bar colouring
  var percentInt = parseInt(rpmBar.style.width);
  if (percentInt > 85)
    rpmBar.style.setProperty('background-color', 'red', 'important');
  else if (percentInt > 60)
    rpmBar.style.setProperty('background-color', 'yellow', 'important');
  else
    rpmBar.style.setProperty('background-color', 'green', 'important');
});

socket.on('LapTiming', (data) => {
  var currentMinutes = Math.floor((data % 3600000) / 60000);
  var currentSeconds = (Math.floor((data % 60000) / 1000)).toString().padStart(2, '0');
  var currentMilliseconds = (data % 1000).toString().padStart(3, '0');
  
  var currentLap = document.getElementById('currentLap');

  currentLap.textContent = `${currentMinutes}:${currentSeconds}.${currentMilliseconds}`;
});

socket.on('LapStats', (lastLap, bestLap, pbLap) => {
  var lastMinutes = Math.floor((lastLap % 3600000) / 60000);
  var lastSeconds = (Math.floor((lastLap % 60000) / 1000)).toString().padStart(2, '0');
  var lastMilliseconds = (lastLap % 1000).toString().padStart(3, '0');

  var bestMinutes = Math.floor((bestLap % 3600000) / 60000);
  var bestSeconds = (Math.floor((bestLap % 60000) / 1000)).toString().padStart(2, '0');
  var bestMilliseconds = (bestLap % 1000).toString().padStart(3, '0');

  var pbMinutes = Math.floor((pbLap % 3600000) / 60000);
  var pbSeconds = (Math.floor((pbLap % 60000) / 1000)).toString().padStart(2, '0');
  var pbMilliseconds = (pbLap % 1000).toString().padStart(3, '0');

  var lastLap = document.getElementById('lastLap');
  var bestLap = document.getElementById('bestLap');
  var pbLap = document.getElementById('pbLap');

  lastLap.textContent = `${lastMinutes}:${lastSeconds}.${lastMilliseconds.to}`;
  bestLap.textContent = `${bestMinutes}:${bestSeconds}.${bestMilliseconds}`;  
  pbLap.textContent = `${pbMinutes}:${pbSeconds}.${pbMilliseconds}`;
});