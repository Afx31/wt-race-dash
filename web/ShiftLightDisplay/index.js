var socket = io.connect('localhost:3000');

let shiftLightRange1 = 0;
let shiftLightRange2 = 0;
let shiftLightRange3 = 0;
let shiftLightRange4 = 0;
let shiftLightRange5 = 0;
let shiftLightRange6 = 0;
let shiftLightRange7 = 0;
// let currentRpm = 0;

document.addEventListener('DOMContentLoaded', () => {
  fetch('/config')
    .then(res => res.json())
    .then(data => {
      shiftLightRange1 = data.shiftLight1,
      shiftLightRange2 = data.shiftLight2,
      shiftLightRange3 = data.shiftLight3,
      shiftLightRange4 = data.shiftLight4,
      shiftLightRange5 = data.shiftLight5,
      shiftLightRange6 = data.shiftLight6,
      shiftLightRange7 = data.shiftLight7
    })
    .catch(error => {
        console.error('Error fetching the config:', error);
    });
});

document.addEventListener('DOMContentLoaded', () => {
  socket.on('CANBusMessageDataLogging', (data) => {
    var dataLoggingAlert = document.getElementById('datalogging-alert');
    
    if (data)
      dataLoggingAlert.style.setProperty('background-color', 'red');
    else
      dataLoggingAlert.style.setProperty('background-color', '');
  });
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
  // var tps = document.getElementById('tps');
  // var map = document.getElementById('map');
  var lambdaRatio = document.getElementById('lambdaRatio');
  var oilTemp = document.getElementById('oilTemp');
  var oilPressure = document.getElementById('oilPressure');
  
  var shiftLight1 = document.getElementById('shift-light-1');
  var shiftLight2 = document.getElementById('shift-light-2');
  var shiftLight3 = document.getElementById('shift-light-3');
  var shiftLight4 = document.getElementById('shift-light-4');
  var shiftLight5 = document.getElementById('shift-light-5');
  var shiftLight6 = document.getElementById('shift-light-6');
  var shiftLight7 = document.getElementById('shift-light-7');

  // function animateRpmBar() {
  //   rpmBar.style.width = ((currentRpm / 9000) * 100) + '%';
  //   requestAnimationFrame(animateRpmBar);
  // }

  socket.on('CANBusMessage', (data) => {
    if (data.rpm < shiftLightRange1) { shiftLight1.style.setProperty('background-color', ''); }
    if (data.rpm >= shiftLightRange1) { shiftLight1.style.setProperty('background-color', 'blue'); }
    
    if (data.rpm < shiftLightRange2) { shiftLight2.style.setProperty('background-color', ''); }
    if (data.rpm >= shiftLightRange2) { shiftLight2.style.setProperty('background-color', 'blue'); }

    if (data.rpm < shiftLightRange3) {shiftLight3.style.setProperty('background-color', ''); }
    if (data.rpm >= shiftLightRange3) { shiftLight3.style.setProperty('background-color', 'green'); }

    if (data.rpm < shiftLightRange4) { shiftLight4.style.setProperty('background-color', ''); }
    if (data.rpm >= shiftLightRange4) { shiftLight4.style.setProperty('background-color', 'green'); }

    if (data.rpm < shiftLightRange5) { shiftLight5.style.setProperty('background-color', ''); }
    if (data.rpm >= shiftLightRange5) { shiftLight5.style.setProperty('background-color', 'yellow'); }

    if (data.rpm < shiftLightRange6) { shiftLight6.style.setProperty('background-color', ''); }
    if (data.rpm >= shiftLightRange6) { shiftLight6.style.setProperty('background-color', 'yellow'); }

    if (data.rpm < shiftLightRange7) { shiftLight7.style.setProperty('background-color', ''); }
    if (data.rpm >= shiftLightRange7) { shiftLight7.style.setProperty('background-color', 'red'); }

    // Assign data to UI controls
    rpmBar.style.width = ((data.rpm / 9000) * 100) + '%';
    // currentRpm = data.rpm;

    tpsBar.style.height = data.tps + '%';
    rpmNum.textContent = data.rpm;
    speed.textContent = data.speed;
    gear.textContent = data.gear;
    voltage.textContent = data.voltage;
    iat.textContent = data.iat;
    ect.textContent = data.ect;
    // tps.textContent = data.tps;
    // map.textContent = data.map;
    lambdaRatio.textContent = data.lambdaRatio;
    oilTemp.textContent = data.oilTemp;
    oilPressure.textContent = data.oilPressure;
  });

  // requestAnimationFrame(animateRpmBar);
});
