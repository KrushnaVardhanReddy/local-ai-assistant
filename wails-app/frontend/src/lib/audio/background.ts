import { Capacitor } from '@capacitor/core';
import { BackgroundTask } from '@capawesome/capacitor-background-task';

let taskId: string | null = null;

/**
 * Registers a background task to keep the app (and specifically the microphone instance)
 * active when the screen is locked or the app is pushed to the background.
 */
export async function registerBackgroundAudioTask(): Promise<void> {
    if (!Capacitor.isNativePlatform()) {
        console.warn('Background task is only available on native platforms.');
        return;
    }

    try {
        taskId = await BackgroundTask.beforeExit(async () => {
            // This callback is executed when the app goes into the background.
            // The task will remain active, allowing continuous listening.
            console.log('Background task started.');

            // Note: The actual microphone keeping-alive logic will rely on this task
            // delaying the suspension of the app. The microphone processing loop
            // will continue to run as long as this task is not finished.
        });
        console.log(`Registered background task with ID: ${taskId}`);
    } catch (error) {
        console.error('Failed to register background task:', error);
    }
}

/**
 * Stops the registered background task, allowing the system to suspend the app.
 */
export async function stopBackgroundAudioTask(): Promise<void> {
    if (!taskId) {
        return;
    }

    try {
        await BackgroundTask.finish({ taskId });
        console.log(`Finished background task with ID: ${taskId}`);
        taskId = null;
    } catch (error) {
        console.error('Failed to finish background task:', error);
    }
}
