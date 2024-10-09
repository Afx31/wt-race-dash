const socket = new WebSocket('ws://' + document.location.host + '/ws');

socket.onopen = function(e) {
  console.log('Connected to WebSocket server.');
};
socket.onclose = function (e) {
  console.log('Connection closed');
};
window.addEventListener("beforeunload", function () {
  socket.close();
});

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

// document.addEventListener('DOMContentLoaded', () => {
//   socket.on('CANBusMessageDataLogging', (data) => {
//     var dataLoggingAlert = document.getElementById('datalogging-alert');
    
//     if (data)
//       dataLoggingAlert.style.setProperty('background-color', 'red');
//     else
//       dataLoggingAlert.style.setProperty('background-color', '');
//   });
// });

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

  socket.onmessage = function (event) {
    const data = JSON.parse(event.data);

    // switch(data.Type) {
    //   case 1:
        const rpm = data.Rpm;
        if (rpm < shiftLightRange1) { shiftLight1.style.setProperty('background-color', ''); }
        if (rpm >= shiftLightRange1) { shiftLight1.style.setProperty('background-color', 'blue'); }
        
        if (rpm < shiftLightRange2) { shiftLight2.style.setProperty('background-color', ''); }
        if (rpm >= shiftLightRange2) { shiftLight2.style.setProperty('background-color', 'blue'); }

        if (rpm < shiftLightRange3) {shiftLight3.style.setProperty('background-color', ''); }
        if (rpm >= shiftLightRange3) { shiftLight3.style.setProperty('background-color', 'green'); }

        if (rpm < shiftLightRange4) { shiftLight4.style.setProperty('background-color', ''); }
        if (rpm >= shiftLightRange4) { shiftLight4.style.setProperty('background-color', 'green'); }

        if (rpm < shiftLightRange5) { shiftLight5.style.setProperty('background-color', ''); }
        if (rpm >= shiftLightRange5) { shiftLight5.style.setProperty('background-color', 'yellow'); }

        if (rpm < shiftLightRange6) { shiftLight6.style.setProperty('background-color', ''); }
        if (rpm >= shiftLightRange6) { shiftLight6.style.setProperty('background-color', 'yellow'); }

        if (rpm < shiftLightRange7) { shiftLight7.style.setProperty('background-color', ''); }
        if (rpm >= shiftLightRange7) { shiftLight7.style.setProperty('background-color', 'red'); }

        // Assign data to UI controls
        rpmBar.style.width = ((rpm / 9000) * 100) + '%';
        // currentRpm = rpm;

        tpsBar.style.height = data.Tps + '%';
        rpmNum.textContent = rpm;
        speed.textContent = data.Speed;
        gear.textContent = data.Gear;
        voltage.textContent = data.Voltage;
        iat.textContent = data.Iat;
        ect.textContent = data.Ect;
        // tps.textContent = data.Tps;
        // map.textContent = data.Map;
        lambdaRatio.textContent = data.LambdaRatio;
        oilTemp.textContent = data.OilTemp;
        oilPressure.textContent = data.OilPressure;
      // case 2:
      //   console.log('--------------------------------------')
      //   console.log(data.Time)
      
    // }
  };

  // requestAnimationFrame(animateRpmBar);
});
