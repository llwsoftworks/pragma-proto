import { writable } from 'svelte/store';

export type ToastType = 'success' | 'error' | 'info' | 'warning';

export interface Toast {
	id: string;
	type: ToastType;
	message: string;
	/** Optional undo callback — shown for 10 seconds per spec. */
	onUndo?: () => void;
	duration: number;
}

const { subscribe, update } = writable<Toast[]>([]);

function addToast(type: ToastType, message: string, onUndo?: () => void, duration = 4000) {
	const id = crypto.randomUUID();
	update((toasts) => [...toasts, { id, type, message, onUndo, duration }]);

	setTimeout(() => {
		update((toasts) => toasts.filter((t) => t.id !== id));
	}, onUndo ? 10000 : duration); // 10s for undo toasts per spec

	return id;
}

export const notifications = {
	subscribe,
	success: (message: string, onUndo?: () => void) => addToast('success', message, onUndo),
	error: (message: string) => addToast('error', message),
	info: (message: string) => addToast('info', message),
	warning: (message: string) => addToast('warning', message),
	dismiss: (id: string) => update((toasts) => toasts.filter((t) => t.id !== id))
};
