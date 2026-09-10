export const isBrowser = typeof window !== 'undefined';

export const checkIsLinkInternal = (href: string): boolean => href[0] == '/';
