chrome.runtime.onInstalled.addListener(() => {
  chrome.storage.local.set({ backend_url: 'http://127.0.0.1:8765' }, () => {
    console.log('BarnOwl AI extension installed. Default backend URL set to http://127.0.0.1:8765');
  });
});