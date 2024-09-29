var socket = io.connect('localhost:3000');

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

  socket.on('CANBusMessage', (data) => {
    rpmBar.style.width = ((data.rpm / 9000) * 100) + '%';
    rpmNum.textContent = data.rpm;
    speed.textContent = data.speed;
    voltage.textContent = data.voltage;
    iat.textContent = data.iat;
    ect.textContent = data.ect;
    tps.textContent = data.tps;
    map.textContent = data.map;
    inj.textContent = data.inj;
    ign.textContent = data.ign;
    lambdaRatio.textContent = data.lambdaRatio;
    oilTemp.textContent = data.oilTemp;
    oilPressure.textContent = data.oilPressure;
  });
});


// ------------------------------ OLD SETUP BELOW ------------------------------

// socket.on('DataLogging', (data) => {
//   var speed = document.getElementById('speed');
//   var voltage = document.getElementById('voltage');
//   var iat = document.getElementById('iat');
//   var ect = document.getElementById('ect');
//   var tps = document.getElementById('tps');
//   var map = document.getElementById('map');
//   var lambdaRatio = document.getElementById('lambdaRatio');
//   var oilTemp = document.getElementById('oilTemp');
//   var oilPressure = document.getElementById('oilPressure');

//   speed.textContent = data.speed;
//   voltage.textContent = (data.voltage / 10).toFixed(1); //TODO: This needs to be moved to backend
//   iat.textContent = data.iat;
//   ect.textContent = data.ect;
//   tps.textContent = data.tps;
//   map.textContent = (data.map / 10) / 2;
//   lambdaRatio.textContent = (32768 / data.lambdaRatio).toFixed(2); //TODO: This needs to be moved to backend
//   oilTemp.textContent = data.oilTemp;
//   oilPressure.textContent = data.oilPressure;
// });

// socket.on('CANBusMessage', (data) => {
//   // if (data.changeDisplay === 1)
//   //   window.location.href = 'http://localhost:3000/';

//   var rpmBar = document.getElementById('rpmbar');
//   var rpmNum = document.getElementById('rpmNum');

//   // RPM progressive bar
//   rpmBar.style.setProperty('max-width', '1920px', 'important'); //1582px
//   var rpmbarPercentage = (data.rpm / 9000) * 100; // = (currentRpm/redlineRpm) * 100

//   // Assign data to UI controls
//   rpmBar.style.width = `${rpmbarPercentage}%`;
//   rpmNum.textContent = data.rpm;
  
//   // RPM Bar colouring
//   var percentInt = parseInt(rpmBar.style.width);
//   if (percentInt > 85)
//     rpmBar.style.setProperty('background-color', 'red', 'important');
//   else if (percentInt > 60)
//     rpmBar.style.setProperty('background-color', 'yellow', 'important');
//   else
//     rpmBar.style.setProperty('background-color', 'green', 'important');
// });