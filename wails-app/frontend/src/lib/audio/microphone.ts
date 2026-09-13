import { apiFetch } from "../api";
import { Capacitor } from '@capacitor/core';
import { Microphone } from '@capacitor-community/microphone';

export class AudioRecorder {
  private mediaRecorder: MediaRecorder | null = null;
  private audioChunks: Blob[] = [];

  async requestPermissions(): Promise<boolean> {
    if (Capacitor.isNativePlatform()) {
      try {
        const permission = await Microphone.checkPermissions();
        if (permission.microphone === 'granted') {
          return true;
        }
        const result = await Microphone.requestPermissions();
        return result.microphone === 'granted';
      } catch (error) {
        console.error("Native microphone permission error:", error);
        return false;
      }
    } else {
      try {
        const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
        stream.getTracks().forEach(track => track.stop());
        return true;
      } catch (error) {
        console.error("Web microphone permission error:", error);
        return false;
      }
    }
  }

  async startRecording(): Promise<void> {
    if (Capacitor.isNativePlatform()) {
      await Microphone.startRecording();
    } else {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      this.mediaRecorder = new MediaRecorder(stream);
      this.audioChunks = [];

      this.mediaRecorder.ondataavailable = (event) => {
        if (event.data.size > 0) {
          this.audioChunks.push(event.data);
        }
      };

      this.mediaRecorder.start();
    }
  }

  async stopRecording(): Promise<Blob | null> {
    if (Capacitor.isNativePlatform()) {
      try {
        const result = await Microphone.stopRecording();
        if (result.webPath) {
            const response = await apiFetch(result.webPath);
            const blob = await response.blob();
            return blob;
        }
        return null;
      } catch (error) {
        console.error("Native microphone stop recording error:", error);
        return null;
      }
    } else {
      return new Promise((resolve) => {
        if (!this.mediaRecorder) {
          resolve(null);
          return;
        }

        this.mediaRecorder.onstop = () => {
          const blob = new Blob(this.audioChunks, { type: 'audio/webm' });
          resolve(blob);

          if (this.mediaRecorder && this.mediaRecorder.stream) {
              this.mediaRecorder.stream.getTracks().forEach(track => track.stop());
          }
          this.mediaRecorder = null;
          this.audioChunks = [];
        };

        this.mediaRecorder.stop();
      });
    }
  }
}
