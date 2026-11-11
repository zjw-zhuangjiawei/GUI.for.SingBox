// Types for the v2-compatible runtime shim in runtime.js.
// Only the functions this app actually calls are declared: a new call site
// then fails type-checking instead of blowing up at runtime.
export interface Position {
    x: number;
    y: number;
}

export interface Size {
    w: number;
    h: number;
}

export interface NotificationOptions {
    id: string;
    title: string;
    subtitle?: string;
    body?: string;
    categoryId?: string;
    data?: { [key: string]: any };
}

export function EventsOn(name: string, callback: (...data: any) => void): () => void;

export function EventsOff(...names: string[]): void;

export function EventsEmit(name: string, ...data: any): void;

export function BrowserOpenURL(url: string): void;

export function ClipboardGetText(): Promise<string>;

export function ClipboardSetText(text: string): Promise<boolean>;

export function Show(): void;

export function WindowHide(): void;

export function WindowMinimise(): void;

export function WindowToggleMaximise(): void;

export function WindowIsMaximised(): Promise<boolean>;

export function WindowIsMinimised(): Promise<boolean>;

export function WindowIsFullscreen(): Promise<boolean>;

export function WindowSetSize(width: number, height: number): void;

export function WindowGetSize(): Promise<Size>;

export function WindowSetPosition(x: number, y: number): void;

export function WindowGetPosition(): Promise<Position>;

export function WindowSetAlwaysOnTop(alwaysOnTop: boolean): void;

export function WindowReloadApp(): void;

export function WindowSetSystemDefaultTheme(): void;

export function IsNotificationAvailable(): Promise<boolean>;

export function RequestNotificationAuthorization(): Promise<boolean>;

export function SendNotification(options: NotificationOptions): Promise<void>;
