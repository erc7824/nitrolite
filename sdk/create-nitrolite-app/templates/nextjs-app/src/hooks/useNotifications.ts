import type { TransferNotificationResponseParams } from '@erc7824/nitrolite';
import NotificationStore from '../store/NotificationStore';
import { useEffect, useRef } from 'react';
import { type Transaction as Notification } from '@erc7824/nitrolite';
import { useSnapshot } from 'valtio';

export const handleTransferNotification = (params: TransferNotificationResponseParams) => {
    NotificationStore.addNotifications(params);
};

export const useNotifications = (onNotification?: (n: Notification) => void) => {
    const { notificationTimeout, notifications } = useSnapshot(NotificationStore.state);
    const timersRef = useRef<Map<number, ReturnType<typeof setTimeout>>>(new Map());

    useEffect(() => {
        Object.values(notifications).forEach((notification) => {
            if (timersRef.current.has(notification.id)) {
                return;
            }

            const timer = setTimeout(() => {
                NotificationStore.dropNotification(notification.id);
                timersRef.current.delete(notification.id);
            }, notificationTimeout);

            timersRef.current.set(notification.id, timer);
            onNotification?.(notification);
        });
    }, [notifications]);

    // Hook destructor to clear timers when the component unmounts
    useEffect(() => {
        return () => {
            timersRef.current.forEach((timer) => clearTimeout(timer));
            timersRef.current.clear();
        };
    }, []);
};
