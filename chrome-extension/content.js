// content.js
console.log('BarnOwl AI content script loaded');

let sidebarOpen = false;
let iframe = null;

function injectSidebar() {
  iframe = document.createElement('iframe');
  iframe.src = chrome.runtime.getURL('sidebar.html');
  iframe.style.position = 'fixed';
  iframe.style.top = '0';
  iframe.style.right = '0';
  iframe.style.width = '360px';
  iframe.style.height = '100vh';
  iframe.style.border = 'none';
  iframe.style.zIndex = '999999';
  iframe.style.display = 'none'; // hidden by default
  iframe.style.boxShadow = '-4px 0 15px rgba(0, 0, 0, 0.5)';
  document.body.appendChild(iframe);
}

function injectButton() {
  const button = document.createElement('button');
  button.innerText = '🦉 BarnOwl';
  button.style.position = 'fixed';
  button.style.bottom = '16px';
  button.style.right = '16px';
  button.style.zIndex = '1000000';
  button.style.padding = '10px 16px';
  button.style.backgroundColor = '#1a1a2e';
  button.style.color = 'white';
  button.style.border = '2px solid #7c3aed';
  button.style.borderRadius = '24px';
  button.style.fontWeight = 'bold';
  button.style.cursor = 'pointer';
  button.style.boxShadow = '0 4px 6px rgba(0,0,0,0.3)';

  button.addEventListener('click', () => {
    sidebarOpen = !sidebarOpen;
    iframe.style.display = sidebarOpen ? 'block' : 'none';
    button.style.right = sidebarOpen ? '376px' : '16px';
  });

  document.body.appendChild(button);
}

let captionBuffer = [];
let debounceTimer = null;
let currentSpeaker = 'Unknown';

function sendCaptions() {
  if (captionBuffer.length > 0 && iframe && iframe.contentWindow) {
    // Google Meet sometimes repeats text, or we capture it piece by piece.
    // We join it together to send.
    const text = captionBuffer.join(' ');
    iframe.contentWindow.postMessage({
      type: 'BARNOWL_CAPTION',
      text: text,
      speaker: currentSpeaker
    }, '*');
    captionBuffer = [];
  }
}

function setupMeetObserver() {
    // Use MutationObserver on the DOM to detect new captions
    const observer = new MutationObserver((mutations) => {
        let newCaptions = false;

        mutations.forEach(mutation => {
            if (mutation.type === 'childList') {
                mutation.addedNodes.forEach(node => {
                    if (node.nodeType === Node.ELEMENT_NODE) {

                        // Extract the speaker name and text from added nodes (using common Google Meet DOM heuristics)
                        // .CNusmb is commonly used for the speaker's name
                        const speakerEl = node.classList && node.classList.contains('CNusmb') ? node : node.querySelector('.CNusmb, .zs7s8d');
                        if (speakerEl && speakerEl.textContent) {
                            currentSpeaker = speakerEl.textContent.trim();
                        }

                        // .iTTPOb is commonly used for caption text chunks
                        const textEl = node.classList && node.classList.contains('iTTPOb') ? node : node.querySelector('.iTTPOb, .Fq31v');
                        if (textEl && textEl.textContent) {
                            const text = textEl.textContent.trim();
                            if (text) {
                                captionBuffer.push(text);
                                newCaptions = true;
                            }
                        } else if (node.classList && (node.classList.contains('Mz6pfc') || node.classList.contains('YTbUzc'))) {
                             // Alternative caption node structures
                             const text = node.textContent.trim();
                             if (text) {
                                 captionBuffer.push(text);
                                 newCaptions = true;
                             }
                        }
                    }
                });
            }
        });

        if (newCaptions) {
            // Debounce captions (collect text for 2 seconds before sending)
            clearTimeout(debounceTimer);
            debounceTimer = setTimeout(sendCaptions, 2000);
        }
    });

    observer.observe(document.body, {
        childList: true,
        subtree: true
    });
}

function init() {
  injectSidebar();
  injectButton();
  setupMeetObserver();
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', init);
} else {
  init();
}
