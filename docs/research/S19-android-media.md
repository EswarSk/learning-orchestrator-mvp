# Android — MediaSessionManager

Source: https://developer.android.com/reference/android/media/session/MediaSessionManager
Retrieved: September 17, 2026 (Pacific time)

[Skip to main content](https://developer.android.com/reference/android/media/session/MediaSessionManager#main-content)

[![Android Developers](https://www.gstatic.com/devrel-devsite/prod/veec7311b6c5f99ef32994eb65aab7022195f72bfa4f3d934bdc7556da7fa7c3b/android/images/lockup.png)](https://developer.android.com/)

`/`

Language

- [English](https://developer.android.com/reference/android/media/session/MediaSessionManager)
- [Deutsch](https://developer.android.com/reference/android/media/session/MediaSessionManager?hl=de)
- [Español – América Latina](https://developer.android.com/reference/android/media/session/MediaSessionManager?hl=es-419)
- [Français](https://developer.android.com/reference/android/media/session/MediaSessionManager?hl=fr)
- [Indonesia](https://developer.android.com/reference/android/media/session/MediaSessionManager?hl=id)
- [Polski](https://developer.android.com/reference/android/media/session/MediaSessionManager?hl=pl)
- [Português – Brasil](https://developer.android.com/reference/android/media/session/MediaSessionManager?hl=pt-br)
- [Tiếng Việt](https://developer.android.com/reference/android/media/session/MediaSessionManager?hl=vi)
- [中文 – 简体](https://developer.android.com/reference/android/media/session/MediaSessionManager?hl=zh-cn)
- [日本語](https://developer.android.com/reference/android/media/session/MediaSessionManager?hl=ja)
- [한국어](https://developer.android.com/reference/android/media/session/MediaSessionManager?hl=ko)

[Android Studio](https://developer.android.com/studio)Sign in

- [API reference](https://developer.android.com/reference)

- [Android Developers](https://developer.android.com/)
- [Develop](https://developer.android.com/develop)
- [API reference](https://developer.android.com/reference)


Added in [API level 21](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

Summary:

[Nested Classes](https://developer.android.com/reference/android/media/session/MediaSessionManager#nestedclasses)


\| [Methods](https://developer.android.com/reference/android/media/session/MediaSessionManager#pubmethods)


\| [Inherited Methods](https://developer.android.com/reference/android/media/session/MediaSessionManager#inhmethods)

# MediaSessionManager    Stay organized with collections      Save and categorize content based on your preferences.

* * *

[Kotlin](https://developer.android.com/reference/kotlin/android/media/session/MediaSessionManager "View this page in Kotlin") \|Java

`
public

final

class
MediaSessionManager
`

`

    extends Object

``

`

|     |     |
| --- | --- |
| [java.lang.Object](https://developer.android.com/reference/java/lang/Object) |
| ↳ | android.media.session.MediaSessionManager |

* * *

Provides support for interacting with `media sessions`
that applications have published to express their ongoing media playback
state.

**See also:**

- `MediaSession`
- `MediaController`

## Summary

| ### Nested classes |
| --- |
| `<br>        interface` | `MediaSessionManager.OnActiveSessionsChangedListener`<br>Listens for changes to the list of active sessions. |
| `<br>        interface` | `MediaSessionManager.OnMediaKeyEventSessionChangedListener`<br>Listener to receive changes in the media key event session, which would receive a media key<br>event unless specified. |
| `<br>        interface` | `MediaSessionManager.OnSession2TokensChangedListener`<br>_This interface was deprecated_<br>_in API level 37._<br>_`MediaSession2` is deprecated._ |
| `<br>        class` | `MediaSessionManager.RemoteUserInfo`<br>Information of a remote user of `MediaSession` or `MediaBrowserService`. |

| ### Public methods |
| --- |
| `<br>        void` | `<br>      addOnActiveSessionsChangedListener(MediaSessionManager.OnActiveSessionsChangedListener sessionListener, ComponentName notificationListener)<br>`<br>Add a listener to be notified when the list of active sessions changes. |
| `<br>        void` | `<br>      addOnActiveSessionsChangedListener(MediaSessionManager.OnActiveSessionsChangedListener sessionListener, ComponentName notificationListener, Handler handler)<br>`<br>Add a listener to be notified when the list of active sessions changes. |
| `<br>        void` | `<br>      addOnActiveSessionsForPackageChangedListener(String packageName, Executor executor, ComponentName notificationListener, MediaSessionManager.OnActiveSessionsChangedListener sessionListener)<br>`<br>Add a listener to be notified when the list of active controllers changes for the `packageName`, as returned by `getActiveSessionsForPackage(String)`. |
| `<br>        void` | `<br>      addOnActiveSessionsForPackageChangedListener(String packageName, Executor executor, MediaSessionManager.OnActiveSessionsChangedListener sessionListener)<br>`<br>Add a listener to be notified when the list of active sessions changes for the `packageName`, as returned by `getActiveSessionsForPackage(String)`. |
| `<br>        void` | `<br>      addOnMediaKeyEventSessionChangedListener(Executor executor, MediaSessionManager.OnMediaKeyEventSessionChangedListener listener)<br>`<br>Add a listener to be notified when the media key session is changed. |
| `<br>        void` | `<br>      addOnSession2TokensChangedListener(MediaSessionManager.OnSession2TokensChangedListener listener)<br>`<br>_This method was deprecated_<br>_in API level 37._<br>_`MediaSession2` is deprecated._ |
| `<br>        void` | `<br>      addOnSession2TokensChangedListener(MediaSessionManager.OnSession2TokensChangedListener listener, Handler handler)<br>`<br>_This method was deprecated_<br>_in API level 37._<br>_`MediaSession2` is deprecated._ |
| `<br>        List<MediaController>` | `<br>      getActiveSessions(ComponentName notificationListener)<br>`<br>Get a list of controllers for all active sessions. |
| `<br>        List<MediaSession.Token>` | `<br>      getActiveSessionsForPackage(String packageName)<br>`<br>Get a list of `MediaSession.Token` for all active sessions for the `packageName`. |
| `<br>        List<MediaSession.Token>` | `<br>      getActiveSessionsForPackage(String packageName, ComponentName notificationListener)<br>`<br>Get a list of `MediaSession.Token` for all active sessions for the `packageName`. |
| `<br>        MediaSession.Token` | `<br>      getMediaKeyEventSession()<br>`<br>Gets the media key event session, which would receive a media key event unless specified. |
| `<br>        String` | `<br>      getMediaKeyEventSessionPackageName()<br>`<br>Gets the package name of the media key event session. |
| `<br>        List<Session2Token>` | `<br>      getSession2Tokens()<br>`<br>_This method was deprecated_<br>_in API level 37._<br>_`MediaSession2` is deprecated._ |
| `<br>        boolean` | `<br>      isTrustedForMediaControl(MediaSessionManager.RemoteUserInfo userInfo)<br>`<br>Checks whether the remote user is a trusted app. |
| `<br>        void` | `<br>      notifySession2Created(Session2Token token)<br>`<br>_This method was deprecated_<br>_in API level 31._<br>_Don't use this method. A new media session is notified automatically._ |
| `<br>        void` | `<br>      removeOnActiveSessionsChangedListener(MediaSessionManager.OnActiveSessionsChangedListener sessionListener)<br>`<br>Stop receiving active sessions updates on the specified listener. |
| `<br>        void` | `<br>      removeOnActiveSessionsForPackageChangedListener(MediaSessionManager.OnActiveSessionsChangedListener sessionListener)<br>`<br>Stop receiving active sessions updates on the specified listener for package. |
| `<br>        void` | `<br>      removeOnMediaKeyEventSessionChangedListener(MediaSessionManager.OnMediaKeyEventSessionChangedListener listener)<br>`<br>Stop receiving updates on media key event session change on the specified listener. |
| `<br>        void` | `<br>      removeOnSession2TokensChangedListener(MediaSessionManager.OnSession2TokensChangedListener listener)<br>`<br>_This method was deprecated_<br>_in API level 37._<br>_`MediaSession2` is deprecated._ |

| ### Inherited methods |
| --- |
| From class
`

          java.lang.Object

`

|     |     |
| --- | --- |
| `<br>        Object` | `<br>      clone()<br>`<br>Creates and returns a copy of this object. |
| `<br>        boolean` | `<br>      equals(Object obj)<br>`<br>Indicates whether some other object is "equal to" this one. |
| `<br>        void` | `<br>      finalize()<br>`<br>Called by the garbage collector on an object when garbage collection<br>determines that there are no more references to the object. |
| `<br>        final<br>        Class<?>` | `<br>      getClass()<br>`<br>Returns the runtime class of this `Object`. |
| `<br>        int` | `<br>      hashCode()<br>`<br>Returns a hash code value for the object. |
| `<br>        final<br>        void` | `<br>      notify()<br>`<br>Wakes up a single thread that is waiting on this object's<br>monitor. |
| `<br>        final<br>        void` | `<br>      notifyAll()<br>`<br>Wakes up all threads that are waiting on this object's monitor. |
| `<br>        String` | `<br>      toString()<br>`<br>Returns a string representation of the object. |
| `<br>        final<br>        void` | `<br>      wait(long timeoutMillis, int nanos)<br>`<br>Causes the current thread to wait until it is awakened, typically<br>by being _notified_ or _interrupted_, or until a<br>certain amount of real time has elapsed. |
| `<br>        final<br>        void` | `<br>      wait(long timeoutMillis)<br>`<br>Causes the current thread to wait until it is awakened, typically<br>by being _notified_ or _interrupted_, or until a<br>certain amount of real time has elapsed. |
| `<br>        final<br>        void` | `<br>      wait()<br>`<br>Causes the current thread to wait until it is awakened, typically<br>by being _notified_ or _interrupted_. | |

## Public methods

### addOnActiveSessionsChangedListener

Added in [API level 21](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

```
public void addOnActiveSessionsChangedListener (MediaSessionManager.OnActiveSessionsChangedListener sessionListener,
                ComponentName notificationListener)
```

Add a listener to be notified when the list of active sessions changes.

This requires the `Manifest.permission.MEDIA_CONTENT_CONTROL` permission be
held by the calling app. You may also retrieve this list if your app is an enabled
notificationlistener using the `NotificationListenerService` APIs, in which case you
must pass the `ComponentName` of your enabled listener.

| Parameters |
| --- |
| `sessionListener` | `MediaSessionManager.OnActiveSessionsChangedListener`: The listener to add.<br>This value cannot be `null`. |
| `notificationListener` | `ComponentName`: The enabled notification listener component. May be null. |

### addOnActiveSessionsChangedListener

Added in [API level 21](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

```
public void addOnActiveSessionsChangedListener (MediaSessionManager.OnActiveSessionsChangedListener sessionListener,
                ComponentName notificationListener,
                Handler handler)
```

Add a listener to be notified when the list of active sessions changes.

This requires the `Manifest.permission.MEDIA_CONTENT_CONTROL` permission be
held by the calling app. You may also retrieve this list if your app is an enabled
notification listener using the `NotificationListenerService` APIs, in which case you
must pass the `ComponentName` of your enabled listener. Updates will be posted to the
handler specified or to the caller's thread if the handler is null.

| Parameters |
| --- |
| `sessionListener` | `MediaSessionManager.OnActiveSessionsChangedListener`: The listener to add.<br>This value cannot be `null`. |
| `notificationListener` | `ComponentName`: The enabled notification listener component. May be null. |
| `handler` | `Handler`: The handler to post events to.<br>This value may be `null`. |

### addOnActiveSessionsForPackageChangedListener

Added in [API level 37](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

```
public void addOnActiveSessionsForPackageChangedListener (String packageName,
                Executor executor,
                ComponentName notificationListener,
                MediaSessionManager.OnActiveSessionsChangedListener sessionListener)
```

Add a listener to be notified when the list of active controllers changes for the `packageName`, as returned by `getActiveSessionsForPackage(String)`. The
registered listener can be removed by calling `removeOnActiveSessionsForPackageChangedListener(OnActiveSessionsChangedListener)`.

If the `packageName` is different from the `Context.getPackageName()`, it
requires the `Manifest.permission.MEDIA_CONTENT_CONTROL` permission be held by
the calling app. You may also attach a listener if your app is an enabled notification
listener using the `NotificationListenerService` APIs, in which case you must pass the
`ComponentName` of your enabled listener.

| Parameters |
| --- |
| `packageName` | `String`: The package name filter to be applied for the listener.<br>This value cannot be `null`. |
| `executor` | `Executor`: The executor to be used for sending events to listener.<br>This value cannot be `null`.<br>Callback and listener events are dispatched through this<br> `Executor`, providing an easy way to control which thread is<br> used. To dispatch events through the main thread of your<br> application, you can use<br> `Context.getMainExecutor()`.<br> Otherwise, provide an `Executor` that dispatches to an appropriate thread. |
| `notificationListener` | `ComponentName`: The enabled notification listener component. May be null. |
| `sessionListener` | `MediaSessionManager.OnActiveSessionsChangedListener`: The listener to add.<br>This value cannot be `null`. |

### addOnActiveSessionsForPackageChangedListener

Added in [API level 37](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

```
public void addOnActiveSessionsForPackageChangedListener (String packageName,
                Executor executor,
                MediaSessionManager.OnActiveSessionsChangedListener sessionListener)
```

Add a listener to be notified when the list of active sessions changes for the `packageName`, as returned by `getActiveSessionsForPackage(String)`. The
registered listener can be removed by calling `removeOnActiveSessionsForPackageChangedListener(OnActiveSessionsChangedListener)`.

If the `packageName` is different from the `Context.getPackageName()`, it
requires the `Manifest.permission.MEDIA_CONTENT_CONTROL` permission be held by
the calling app.

| Parameters |
| --- |
| `packageName` | `String`: The package name filter to be applied for the listener.<br>This value cannot be `null`. |
| `executor` | `Executor`: The executor to be used for sending events to listener.<br>This value cannot be `null`.<br>Callback and listener events are dispatched through this<br> `Executor`, providing an easy way to control which thread is<br> used. To dispatch events through the main thread of your<br> application, you can use<br> `Context.getMainExecutor()`.<br> Otherwise, provide an `Executor` that dispatches to an appropriate thread. |
| `sessionListener` | `MediaSessionManager.OnActiveSessionsChangedListener`: The listener to add.<br>This value cannot be `null`. |

### addOnMediaKeyEventSessionChangedListener

Added in [API level 33](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

```
public void addOnMediaKeyEventSessionChangedListener (Executor executor,
                MediaSessionManager.OnMediaKeyEventSessionChangedListener listener)
```

Add a listener to be notified when the media key session is changed.

This requires the `Manifest.permission.MEDIA_CONTENT_CONTROL`
permission be held by the calling app, or the app has an enabled notification listener
using the `NotificationListenerService` APIs. If none of them applies, it will throw
a `SecurityException`.

| Parameters |
| --- |
| `executor` | `Executor`: The executor on which the listener should be invoked.<br>This value cannot be `null`.<br>Callback and listener events are dispatched through this<br> `Executor`, providing an easy way to control which thread is<br> used. To dispatch events through the main thread of your<br> application, you can use<br> `Context.getMainExecutor()`.<br> Otherwise, provide an `Executor` that dispatches to an appropriate thread. |
| `listener` | `MediaSessionManager.OnMediaKeyEventSessionChangedListener`: A `OnMediaKeyEventSessionChangedListener`.<br>This value cannot be `null`. |

### addOnSession2TokensChangedListener

Added in [API level 29](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

Deprecated in
[API level\\
37](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

```
public void addOnSession2TokensChangedListener (MediaSessionManager.OnSession2TokensChangedListener listener)
```

**This method was deprecated**
**in API level 37.**

`MediaSession2` is deprecated.


Adds a listener to be notified when the `getSession2Tokens()` changes.

This API is not generally intended for third party application developers. Apps wanting media
session functionality should use the
[AndroidX Media3\\
Session Library](https://developer.android.com/reference/androidx/media3/session/package-summary).

| Parameters |
| --- |
| `listener` | `MediaSessionManager.OnSession2TokensChangedListener`: The listener to add.<br>This value cannot be `null`. |

### addOnSession2TokensChangedListener

Added in [API level 29](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

Deprecated in
[API level\\
37](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

```
public void addOnSession2TokensChangedListener (MediaSessionManager.OnSession2TokensChangedListener listener,
                Handler handler)
```

**This method was deprecated**
**in API level 37.**

`MediaSession2` is deprecated.


Adds a listener to be notified when the `getSession2Tokens()` changes.

This API is not generally intended for third party application developers. Apps wanting media
session functionality should use the
[AndroidX Media3\\
Session Library](https://developer.android.com/reference/androidx/media3/session/package-summary).

| Parameters |
| --- |
| `listener` | `MediaSessionManager.OnSession2TokensChangedListener`: The listener to add.<br>This value cannot be `null`. |
| `handler` | `Handler`: The handler to call listener on.<br>This value cannot be `null`. |

### getActiveSessions

Added in [API level 21](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

```
public List<MediaController> getActiveSessions (ComponentName notificationListener)
```

Get a list of controllers for all active sessions. The controllers will
be provided in priority order with the most important controller at index
0.

This requires the `Manifest.permission.MEDIA_CONTENT_CONTROL`
permission be held by the calling app. You may also retrieve this list if
your app is an enabled notification listener using the
`NotificationListenerService` APIs, in which case you must pass the
`ComponentName` of your enabled listener.

| Parameters |
| --- |
| `notificationListener` | `ComponentName`: The enabled notification listener component.<br> May be null. |

| Returns |
| --- |
| `List<MediaController>` | A list of controllers for active sessions.<br>This value cannot be `null`. |

### getActiveSessionsForPackage

Added in [API level 37](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

```
public List<MediaSession.Token> getActiveSessionsForPackage (String packageName)
```

Get a list of `MediaSession.Token` for all active sessions for the `packageName`.
The tokens will be provided in priority order with the most important controller at index 0.

If the `packageName` is provided and is different from `Context.getPackageName()`, this requires the `Manifest.permission.MEDIA_CONTENT_CONTROL` permission be held by the calling app.

| Parameters |
| --- |
| `packageName` | `String`: Package name of the application for which tokens should be fetched.<br>This value cannot be `null`. |

| Returns |
| --- |
| `List<MediaSession.Token>` | A list of `MediaSession.Token` for active sessions of the calling application.<br>This value cannot be `null`. |

### getActiveSessionsForPackage

Added in [API level 37](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

```
public List<MediaSession.Token> getActiveSessionsForPackage (String packageName,
                ComponentName notificationListener)
```

Get a list of `MediaSession.Token` for all active sessions for the `packageName`.
The tokens will be provided in priority order with the most important controller at index 0.

If the `packageName` is provided and is different from `Context.getPackageName()`, this requires the `Manifest.permission.MEDIA_CONTENT_CONTROL` permission be held by the calling app. You
may also retrieve this list if your app is an enabled notification listener using the `NotificationListenerService` APIs, in which case you must pass the `ComponentName` of
your enabled listener. If none of them applies, it will throw a `SecurityException`.

| Parameters |
| --- |
| `packageName` | `String`: Package name of the application for which tokens should be fetched.<br>This value cannot be `null`. |
| `notificationListener` | `ComponentName`: The enabled notification listener component. May be null. |

| Returns |
| --- |
| `List<MediaSession.Token>` | A list of controllers for active sessions of the calling application.<br>This value cannot be `null`. |

### getMediaKeyEventSession

Added in [API level 33](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

```
public MediaSession.Token getMediaKeyEventSession ()
```

Gets the media key event session, which would receive a media key event unless specified.

This requires the `Manifest.permission.MEDIA_CONTENT_CONTROL`
permission be held by the calling app, or the app has an enabled notification listener
using the `NotificationListenerService` APIs. If none of them applies, it will throw
a `SecurityException`.

| Returns |
| --- |
| `MediaSession.Token` | The media key event session, which would receive key events by default, unless<br> the caller has specified the target. Can be `null`. |

### getMediaKeyEventSessionPackageName

Added in [API level 33](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

```
public String getMediaKeyEventSessionPackageName ()
```

Gets the package name of the media key event session.

This requires the `Manifest.permission.MEDIA_CONTENT_CONTROL`
permission be held by the calling app, or the app has an enabled notification listener
using the `NotificationListenerService` APIs. If none of them applies, it will throw
a `SecurityException`.

| Returns |
| --- |
| `String` | The package name of the media key event session or the last session's media button<br> receiver if the media key event session is `null`. Returns an empty string<br> if neither of them exists. |

**See also:**

- `getMediaKeyEventSession()`

### getSession2Tokens

Added in [API level 29](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

Deprecated in
[API level\\
37](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

```
public List<Session2Token> getSession2Tokens ()
```

**This method was deprecated**
**in API level 37.**

`MediaSession2` is deprecated.


Gets a list of `Session2Token` with type `Session2Token.TYPE_SESSION` for the
current user.

Although this API can be used without any restriction, each session owners can accept or
reject your uses of `MediaSession2`.

This API is not generally intended for third party application developers. Apps wanting media
session functionality should use the
[AndroidX Media3\\
Session Library](https://developer.android.com/reference/androidx/media3/session/package-summary).

| Returns |
| --- |
| `List<Session2Token>` | A list of `Session2Token`.<br>This value cannot be `null`. |

### isTrustedForMediaControl

Added in [API level 28](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

```
public boolean isTrustedForMediaControl (MediaSessionManager.RemoteUserInfo userInfo)
```

Checks whether the remote user is a trusted app.

An app is trusted if the app holds the
`Manifest.permission.MEDIA_CONTENT_CONTROL` permission or has an enabled
notification listener.

| Parameters |
| --- |
| `userInfo` | `MediaSessionManager.RemoteUserInfo`: The remote user info from either<br> `MediaSession.getCurrentControllerInfo()` or<br> `MediaBrowserService.getCurrentBrowserInfo()`.<br>This value cannot be `null`. |

| Returns |
| --- |
| `boolean` | `true` if the remote user is trusted and its package name matches with the UID.<br> `false` otherwise. |

### notifySession2Created

Added in [API level 29](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

Deprecated in
[API level\\
31](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

```
public void notifySession2Created (Session2Token token)
```

**This method was deprecated**
**in API level 31.**

Don't use this method. A new media session is notified automatically.


Notifies that a new `MediaSession2` with type `Session2Token.TYPE_SESSION` is
created.

Do not use this API directly, but create a new instance through the
`MediaSession2.Builder` instead.

| Parameters |
| --- |
| `token` | `Session2Token`: newly created session2 token.<br>This value cannot be `null`. |

### removeOnActiveSessionsChangedListener

Added in [API level 21](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

```
public void removeOnActiveSessionsChangedListener (MediaSessionManager.OnActiveSessionsChangedListener sessionListener)
```

Stop receiving active sessions updates on the specified listener.

| Parameters |
| --- |
| `sessionListener` | `MediaSessionManager.OnActiveSessionsChangedListener`: The listener to remove.<br>This value cannot be `null`. |

### removeOnActiveSessionsForPackageChangedListener

Added in [API level 37](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

```
public void removeOnActiveSessionsForPackageChangedListener (MediaSessionManager.OnActiveSessionsChangedListener sessionListener)
```

Stop receiving active sessions updates on the specified listener for package.

| Parameters |
| --- |
| `sessionListener` | `MediaSessionManager.OnActiveSessionsChangedListener`: The listener to remove.<br>This value cannot be `null`. |

### removeOnMediaKeyEventSessionChangedListener

Added in [API level 33](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

```
public void removeOnMediaKeyEventSessionChangedListener (MediaSessionManager.OnMediaKeyEventSessionChangedListener listener)
```

Stop receiving updates on media key event session change on the specified listener.

| Parameters |
| --- |
| `listener` | `MediaSessionManager.OnMediaKeyEventSessionChangedListener`: A `OnMediaKeyEventSessionChangedListener`.<br>This value cannot be `null`. |

### removeOnSession2TokensChangedListener

Added in [API level 29](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

Deprecated in
[API level\\
37](https://developer.android.com/guide/topics/manifest/uses-sdk-element#ApiLevels)

```
public void removeOnSession2TokensChangedListener (MediaSessionManager.OnSession2TokensChangedListener listener)
```

**This method was deprecated**
**in API level 37.**

`MediaSession2` is deprecated.


Removes the `OnSession2TokensChangedListener` to stop receiving session token updates.

This API is not generally intended for third party application developers. Apps wanting media
session functionality should use the
[AndroidX Media3\\
Session Library](https://developer.android.com/reference/androidx/media3/session/package-summary).

| Parameters |
| --- |
| `listener` | `MediaSessionManager.OnSession2TokensChangedListener`: The listener to remove.<br>This value cannot be `null`. |

Content and code samples on this page are subject to the licenses described in the [Content License](https://developer.android.com/license). Java and OpenJDK are trademarks or registered trademarks of Oracle and/or its affiliates.

Last updated 2026-08-03 UTC.




\[\[\["Easy to understand","easyToUnderstand","thumb-up"\],\["Solved my problem","solvedMyProblem","thumb-up"\],\["Other","otherUp","thumb-up"\]\],\[\["Missing the information I need","missingTheInformationINeed","thumb-down"\],\["Too complicated / too many steps","tooComplicatedTooManySteps","thumb-down"\],\["Out of date","outOfDate","thumb-down"\],\["Samples / code issue","samplesCodeIssue","thumb-down"\],\["Other","otherDown","thumb-down"\]\],\["Last updated 2026-08-03 UTC."\],\[\],\[\]\]
