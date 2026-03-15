// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

const chatWindow = document.getElementById('chat-window');
const inputText = document.getElementById('input-text');
const btnSend = document.getElementById('btn-send');
const btnRecord = document.getElementById('btn-record');
const iconMic = document.getElementById('icon-mic');
const btnPhone = document.getElementById('btn-phone');
const btnRestart = document.getElementById('btn-restart');

let manualContext = null;
let manualSource = null;
let manualProcessor = null;
let manualStream = null;
let manualChunks = [];
let isManualRecording = false;

let myvad = null;
let isPhoneMode = false;
let currentBotAudio = null;

function float32ToWav(samples, sampleRate) {
    const numChannels = 1;
    const bitsPerSample = 16;
    const byteRate = sampleRate * numChannels * bitsPerSample / 8;
    const blockAlign = numChannels * bitsPerSample / 8;
    const pcmLength = samples.length * 2; 
    const buffer = new ArrayBuffer(44 + pcmLength);
    const view = new DataView(buffer);

    const writeStr = (offset, str) => { for (let i = 0; i < str.length; i++) view.setUint8(offset + i, str.charCodeAt(i)); };
    writeStr(0, 'RIFF');
    view.setUint32(4, 36 + pcmLength, true);
    writeStr(8, 'WAVE');
    writeStr(12, 'fmt ');
    view.setUint32(16, 16, true);          
    view.setUint16(20, 1, true);           
    view.setUint16(22, numChannels, true);
    view.setUint32(24, sampleRate, true);
    view.setUint32(28, byteRate, true);
    view.setUint16(32, blockAlign, true);
    view.setUint16(34, bitsPerSample, true);
    writeStr(36, 'data');
    view.setUint32(40, pcmLength, true);

    let offset = 44;
    for (let i = 0; i < samples.length; i++) {
        const s = Math.max(-1, Math.min(1, samples[i]));
        view.setInt16(offset, s < 0 ? s * 0x8000 : s * 0x7FFF, true);
        offset += 2;
    }

    return new Blob([buffer], { type: 'audio/wav' });
}

function mergeBuffers(buffers) {
    let totalLen = 0;
    for (let i = 0; i < buffers.length; i++) totalLen += buffers[i].length;
    const res = new Float32Array(totalLen);
    let offset = 0;
    for (let i = 0; i < buffers.length; i++) {
        res.set(buffers[i], offset);
        offset += buffers[i].length;
    }
    return res;
}

function appendMessage(sender, type, content) {
    const wrapper = document.createElement('div');
    
    if (type === 'text') {
        let classes = "p-2 ps-3 pe-3 rounded-2 shadow-sm mw-75 text-break ";
        if (sender === 'me') {
            wrapper.className = classes + "align-self-end bg-primary text-white";
						wrapper.textContent = content;
        } else {
            wrapper.className = classes + "align-self-start bg-white border text-dark";
						wrapper.innerHTML = marked.parse(content).trim().replace(/^<p>/, "").replace(/<\/p>$/, "").replace(/<p><\/p>$/,"");
        }
    } else {
        let classes = "mw-75 ";
        if (sender === 'me') {
            wrapper.className = classes + "align-self-end";
        } else {
            wrapper.className = classes + "align-self-start";
        }
        
        const audio = document.createElement('audio');
        audio.controls = true;
        audio.src = content;
        wrapper.appendChild(audio);
        
        if (sender === 'bot' && isPhoneMode) {
            if (currentBotAudio && !currentBotAudio.paused) {
                currentBotAudio.pause();
            }
            currentBotAudio = audio;
            audio.onended = () => { 
                if (currentBotAudio === audio) {
                    currentBotAudio = null;
                }
            };
            
            audio.play().catch(e => {});
        }
    }

    chatWindow.appendChild(wrapper);
    chatWindow.scrollTop = chatWindow.scrollHeight;
}

async function sendText() {
    const text = inputText.value.trim();
    if (!text) return;
    appendMessage('me', 'text', text);
    inputText.value = '';
    try {
        const response = await fetch(window.location.href, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ text: text })
        });
        const data = await response.json();
        appendMessage('bot', 'text', data.text);
    } catch (err) {
        appendMessage('bot', 'text', 'Error');
    }
}

async function sendAudio(blob) {
    const formData = new FormData();
    formData.append("audio", blob, "rec.wav");

    try {
        const response = await fetch(window.location.href, { method: 'POST', body: formData });
        const serverBlob = await response.blob();
        const serverUrl = URL.createObjectURL(serverBlob);
        appendMessage('bot', 'audio', serverUrl);
    } catch (err) {
        appendMessage('bot', 'text', 'Upload failed');
    }
}

btnRecord.addEventListener('click', async () => {
    if (isPhoneMode) { alert("Disable Phone mode first"); return; }
    
    if (!isManualRecording) {
        try {
            manualStream = await navigator.mediaDevices.getUserMedia({ audio: true });
            manualContext = new AudioContext();
            manualSource = manualContext.createMediaStreamSource(manualStream);
            manualProcessor = manualContext.createScriptProcessor(4096, 1, 1);
            manualChunks = [];

            manualProcessor.onaudioprocess = (e) => {
                manualChunks.push(new Float32Array(e.inputBuffer.getChannelData(0)));
            };

            manualSource.connect(manualProcessor);
            manualProcessor.connect(manualContext.destination);

            isManualRecording = true;
            btnRecord.className = "btn btn-warning";
            iconMic.className = "bi bi-stop-circle-fill";
        } catch (e) { alert("Mic Error"); }
    } else {
        if(manualProcessor) {
            manualProcessor.disconnect();
            manualSource.disconnect();
            manualStream.getTracks().forEach(t => t.stop());
            
            const merged = mergeBuffers(manualChunks);
            const wavBlob = float32ToWav(merged, manualContext.sampleRate);
            
            appendMessage('me', 'audio', URL.createObjectURL(wavBlob));
            sendAudio(wavBlob);
            
            manualContext.close();
        }

        isManualRecording = false;
        btnRecord.className = "btn btn-outline-warning";
        iconMic.className = "bi bi-mic-fill";
    }
});

btnPhone.addEventListener('click', async () => {
    if (isManualRecording) return;

    if (!isPhoneMode) {
        btnPhone.disabled = true;
        
        try {
            myvad = await vad.MicVAD.new({
                model: 'v5',
                onnxWASMBasePath: '/pbx/vad/',
                baseAssetPath: '/pbx/vad/',
                
                onSpeechStart: () => {
                    if (currentBotAudio && !currentBotAudio.paused) {
                        console.log("User started speaking: Stopping bot audio.");
                        currentBotAudio.pause();
                        currentBotAudio.currentTime = 0;
                        currentBotAudio = null;
                    }
                },

                onSpeechEnd: (audio) => {
                    const wavBlob = float32ToWav(audio, 16000);
                    const wavUrl = URL.createObjectURL(wavBlob);
                    appendMessage('me', 'audio', wavUrl);
                    sendAudio(wavBlob);
                }
            });

            myvad.start();
            isPhoneMode = true;
            btnPhone.disabled = false;
            btnPhone.className = "btn btn-success btn-blink-active"; 
        } catch (e) {
            btnPhone.disabled = false;
            alert("VAD Error: " + e.message);
        }
    } else {
        if (myvad) {
            myvad.pause();
            myvad = null;
        }
        if (currentBotAudio && !currentBotAudio.paused) {
            currentBotAudio.pause();
            currentBotAudio = null;
        }
        isPhoneMode = false;
        btnPhone.className = "btn btn-outline-secondary";
    }
});

async function checkAudio() {
		if (document.location.search.indexOf("audio=on") > 0) {
						try {
										const stream = await navigator.mediaDevices.getUserMedia({ audio: true });

										stream.getTracks().forEach(track => track.stop());
        
        if (btnPhone) btnPhone.classList.remove('d-none');
        if (btnRecord) btnRecord.classList.remove('d-none');
        
    } catch (err) {
        console.log("Mic access denied or not available.");
    }
	}
}

async function chatRestart() {
  if (confirm("Restart chat?")) {
			inputText.value = "/reset"
			await sendText()
      document.location.href="?";
  }
}

btnRestart.addEventListener('click', chatRestart)
btnSend.addEventListener('click', sendText);
inputText.addEventListener('keypress', (e) => { if (e.key === 'Enter') sendText(); });
checkAudio();

