const fs = require('fs');
const vm = require('vm');

const errors = [];
process.on('uncaughtException', err => errors.push(err));
process.on('unhandledRejection', err => errors.push(err));

// Minimal browser stubs
global.window = {
  addEventListener: () => {},
  requestAnimationFrame: () => {},
};
global.requestAnimationFrame = () => {};

const ctx = {
  save(){}, restore(){}, clearRect(){}, translate(){}, rotate(){},
  beginPath(){}, moveTo(){}, lineTo(){}, arc(){}, closePath(){},
  fill(){}, stroke(){},
  set lineWidth(v){}, set lineCap(v){}, set strokeStyle(v){},
  set fillStyle(v){}, set globalAlpha(v){},
};

global.document = {
  getElementById: () => ({
    clientWidth: 200,
    clientHeight: 200,
    width: 0,
    height: 0,
    getContext: () => ctx,
  }),
};

// Stub fetch to load YAML from disk
global.fetch = () => Promise.resolve({
  text: () => fs.promises.readFile('webos-app/sections.yaml', 'utf8'),
});

const code = fs.readFileSync('webos-app/script.js', 'utf8');

(async () => {
  vm.runInThisContext(code, { filename: 'script.js' });
  await new Promise(r => setTimeout(r, 10));

  if (errors.length) {
    console.error(errors);
    process.exit(1);
  } else {
    console.log('Test passed');
  }
})();
