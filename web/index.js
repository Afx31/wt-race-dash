var socket = io.connect('localhost:3000');

socket.on('CANBusMessage', (data) => {  
  // if (data.changeDisplay === 1)
  //   window.location.href = 'http://localhost:3000/LapTimingDisplay';

  var rpmBar = document.getElementById('rpmbar');
  var rpmNum = document.getElementById('rpmNum');
  var speed = document.getElementById('speed');
  var gear = document.getElementById('gear');
  var voltage = document.getElementById('voltage');
  var iat = document.getElementById('iat');
  var ect = document.getElementById('ect');
  var tpsBar = document.getElementById('tpsbar');
  var tps = document.getElementById('tps');
  var map = document.getElementById('map');
  var lambdaRatio = document.getElementById('lambdaRatio');
  // var inj = document.getElementById('inj');
  // var ign = document.getElementById('ign');
  var oilTemp = document.getElementById('oilTemp');
  var oilPressure = document.getElementById('oilPressure');
  
  // Assign data to UI controls
  rpmBar.style.width = ((data.rpm / 9000) * 100) + '%';

  // if (tpsBar.style.height !== data.tps + '%')
    tpsBar.style.height = data.tps + '%';

  // if (rpmNum.textContent !== data.rpm)
    rpmNum.textContent = data.rpm;
  
  // if (speed.textContent !== data.speed)
    speed.textContent = data.speed;
  
  gear.textContent = data.gear;
  voltage.textContent = data.voltage;  
  iat.textContent = data.iat;
  ect.textContent = data.ect;
  tps.textContent = data.tps;
  map.textContent = data.map;
  lambdaRatio.textContent = data.lambdaRatio;
  // inj.textContent = data.inj;
  // ign.textContent = data.ign;
  oilTemp.textContent = data.oilTemp;
  oilPressure.textContent = data.oilPressure;
});