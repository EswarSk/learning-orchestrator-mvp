# Physical-device acceptance checklist

These checks are intentionally **unrun** until an Apple team, signed development build, approved test regions, provider credentials, and a physical iPhone are available. Simulator or fixture results do not satisfy them.

- Install a signed EAS development build on the named iPhone/iOS version.
- Complete sign-in and an ordinary text lesson before granting Calendar, location, notification, or microphone permission.
- Grant background location; cross an approved region in foreground, background, locked, after reboot, and after network loss. Record OS event and server receipt times.
- Deny and later grant each permission; verify text learning stays usable and prior-account events are cleared after sign-out/account switch.
- Receive a generic push, deep-link into the authenticated opportunity, and confirm private details never appear on the lock screen.
- Start Realtime voice by explicit tap; test route changes, Bluetooth, captions, interruption, reconnect, backgrounding, server hangup, API loss/watchdog, and the three-minute ceiling.
- Record unsupported force-termination and OS-throttling behavior without describing non-delivery as success.
