import { proxy } from 'valtio';
// TODO: It would be good to have separate types for notifications, but BE sends transactions as notifications
import { type Transaction as Notification } from '@erc7824/nitrolite';

export interface INotificationStoreState {
    notifications: Record<number, Notification>;
    notificationTimeout: number; // Timeout in milliseconds for how long notifications should be displayed
}

const state = proxy<INotificationStoreState>({
    notifications: {},
    notificationTimeout: 10000,
});

const NotificationStore = {
    state,
    addNotifications(notifications: Notification[]): void {
        notifications.forEach((notification) => {
            this.addNotification(notification);
        });
    },
    addNotification(notification: Notification): void {
        state.notifications[notification.id] = notification;

        // Auto-dismiss after timeout
        setTimeout(() => {
            this.dropNotification(notification.id);
        }, state.notificationTimeout);
    },
    dropNotification(id: number): void {
        delete state.notifications[id];
    },
    getSortedNotifications(): Notification[] {
        return Object.values(state.notifications).sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());
    },
    clearAllNotifications(): void {
        state.notifications = {};
    },
};

export default NotificationStore;
