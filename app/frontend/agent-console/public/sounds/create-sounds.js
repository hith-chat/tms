// Simple script to create basic alarm sounds using Web Audio API
// Run this in browser console to generate and download sound files

function createAlarmSound(frequency, duration, filename) {
  const audioContext = new (window.AudioContext || window.webkitAudioContext)();
  const sampleRate = audioContext.sampleRate;
  const numChannels = 1;
  const length = sampleRate * duration;
  const buffer = audioContext.createBuffer(numChannels, length, sampleRate);
  const data = buffer.getChannelData(0);
  
  // Generate alarm-like beeping sound
  for (let i = 0; i < length; i++) {
    const time = i / sampleRate;
    const beepPattern = Math.sin(2 * Math.PI * frequency * time) * 
                       (Math.sin(2 * Math.PI * 2 * time) > 0 ? 1 : 0) * // Beeping pattern
                       Math.exp(-time * 0.5); // Fade out
    data[i] = beepPattern * 0.3; // Volume control
  }
  
  // Convert to WAV and download
  const wav = audioBufferToWav(buffer);
  const blob = new Blob([wav], { type: 'audio/wav' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
}

function audioBufferToWav(buffer) {
  const length = buffer.length;
  const arrayBuffer = new ArrayBuffer(44 + length * 2);
  const view = new DataView(arrayBuffer);
  
  // WAV header
  const writeString = (offset, string) => {
    for (let i = 0; i < string.length; i++) {
      view.setUint8(offset + i, string.charCodeAt(i));
    }
  };
  
  writeString(0, 'RIFF');
  view.setUint32(4, 36 + length * 2, true);
  writeString(8, 'WAVE');
  writeString(12, 'fmt ');
  view.setUint32(16, 16, true);
  view.setUint16(20, 1, true);
  view.setUint16(22, 1, true);
  view.setUint32(24, buffer.sampleRate, true);
  view.setUint32(28, buffer.sampleRate * 2, true);
  view.setUint16(32, 2, true);
  view.setUint16(34, 16, true);
  writeString(36, 'data');
  view.setUint32(40, length * 2, true);
  
  // Convert float samples to 16-bit PCM
  const data = buffer.getChannelData(0);
  let offset = 44;
  for (let i = 0; i < length; i++) {
    const sample = Math.max(-1, Math.min(1, data[i]));
    view.setInt16(offset, sample * 0x7FFF, true);
    offset += 2;
  }
  
  return arrayBuffer;
}

// Create the sound files
console.log('Creating alarm sounds...');
createAlarmSound(800, 2, 'alarm-soft.wav');  // 800Hz for 2 seconds
createAlarmSound(1200, 3, 'alarm-urgent.wav'); // 1200Hz for 3 seconds
console.log('Sound files created! Convert to MP3 if needed.');
