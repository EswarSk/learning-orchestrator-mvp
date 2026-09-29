# Expo Location

Source: https://docs.expo.dev/versions/latest/sdk/location/
Retrieved: September 17, 2026, Pacific time.

Hide navigation

Search or Ask AI

`⌘`  `K`

[Home](https://docs.expo.dev/) [Guides](https://docs.expo.dev/guides/overview) [EAS](https://docs.expo.dev/eas) [Reference](https://docs.expo.dev/versions/latest) [Learn](https://docs.expo.dev/tutorial/overview)

Reference version

SDK 57 (latest)

[Archive](https://docs.expo.dev/archive) [Expo Snack](https://snack.expo.dev/) [Discord and forums](https://chat.expo.dev/) [Newsletter](https://expo.dev/mailing-list/signup)

This documentation is available as Markdown for AI agents and LLMs. See the [full Markdown index](https://docs.expo.dev/llms.txt) or append .md to any documentation URL.

# ![Expo Location icon](https://docs.expo.dev/static/images/packages/expo-location.png)Expo Location

A library that provides access to reading geolocation information, polling current location or subscribing location update events from the device.

Android

iOS

Web

Included in Expo Go

Ask AI

[GitHub](https://github.com/expo/expo/tree/sdk-57/packages/expo-location) [npm](https://www.npmjs.com/package/expo-location) [Changelog](https://github.com/expo/expo/tree/sdk-57/packages/expo-location/CHANGELOG.md)

Recommended version:

~57.0.18

Copy page

* * *

`expo-location` allows reading geolocation information from the device. Your app can poll for the current location or subscribe to location update events.

## Installation

Terminal
npm
npmyarnpnpmbun

Copy

`-``npx expo install expo-location`

`-``yarn expo install expo-location`

`-``pnpm expo install expo-location`

`-``bun expo install expo-location`

If you are installing this in an [existing React Native app](https://docs.expo.dev/bare/overview),make sure to [install `expo`](https://docs.expo.dev/bare/installing-expo-modules) in your project.

## Configuration in app config

You can configure `expo-location` using its built-in [config plugin](https://docs.expo.dev/config-plugins/introduction) if you use config plugins in your project ( [Continuous Native Generation (CNG)](https://docs.expo.dev/workflow/continuous-native-generation)). The plugin allows you to configure various properties that cannot be set at runtime and require building a new app binary to take effect. If your app does not use CNG, then you'll need to manually configure the library.

### Example app.json with config plugin

app.json

Copy

```
{
  "expo": {
    "plugins": [\
      [\
        "expo-location",\
        {\
          "locationAlwaysAndWhenInUsePermission": "Allow $(PRODUCT_NAME) to use your location."\
        }\
      ]\
    ]
  }
}
```

### Configurable properties

| Name | Default | Description |
| --- | --- | --- |
| `locationAlwaysAndWhenInUsePermission` | `"Allow $(PRODUCT_NAME) to use your location"` | Only for:<br>iOS<br>A string to set the [`NSLocationAlwaysAndWhenInUseUsageDescription`](https://docs.expo.dev/versions/latest/sdk/location#permission-nslocationalwaysandwheninuseusagedescription) permission message. |
| `locationAlwaysPermission` | `"Allow $(PRODUCT_NAME) to use your location"` | Deprecated<br>• <br>Only for:<br>iOS<br>A string to set the [`NSLocationAlwaysUsageDescription`](https://docs.expo.dev/versions/latest/sdk/location#permission-nslocationalwaysusagedescription) permission message. |
| `locationWhenInUsePermission` | `"Allow $(PRODUCT_NAME) to use your location"` | Only for:<br>iOS<br>A string to set the [`NSLocationWhenInUseUsageDescription`](https://docs.expo.dev/versions/latest/sdk/location#permission-nslocationwheninuseusagedescription) permission message. |
| `motionUsagePermission` | `"Allow $(PRODUCT_NAME) to detect your current motion activity"` | Only for:<br>iOS<br>A string to set the `NSMotionUsageDescription` permission message shown when `getMotionActivityAsync` or `watchMotionActivityAsync` is called for the first time. |
| `isIosBackgroundLocationEnabled` | `false` | Only for:<br>iOS<br>A boolean to enable `location` in the `UIBackgroundModes` in Info.plist. |
| `isAndroidBackgroundLocationEnabled` | `false` | Only for:<br>Android<br>A boolean to enable the [`ACCESS_BACKGROUND_LOCATION`](https://docs.expo.dev/versions/latest/sdk/location#permission-access_background_location) permission. |
| `isAndroidForegroundServiceEnabled` | - | Only for:<br>Android<br>A boolean to enable the [`FOREGROUND_SERVICE`](https://docs.expo.dev/versions/latest/sdk/location#permission-foreground_service) permission and [`FOREGROUND_SERVICE_LOCATION`](https://docs.expo.dev/versions/latest/sdk/location#permission-foreground_service_location). Defaults to `true` if `isAndroidBackgroundLocationEnabled` is `true`, otherwise `false`. |
| `androidForegroundServiceIcon` | - | Only for:<br>Android<br>Local path to an image to use as the icon for the foreground service started by `startLocationUpdatesAsync`. 96x96 all-white png with transparency. If not set, falls back to the `notification_icon` drawable (if configured via `expo-notifications` config plugin), then to the app launcher icon. The launcher icon is full-color and may render as a solid white square since Android requires notification icons to be monochrome. |

Are you using this library in an existing React Native app? [Permalink](https://docs.expo.dev/versions/latest/sdk/location#are-you-using-this-library-in-an)

If you're not using Continuous Native Generation ( [CNG](https://docs.expo.dev/workflow/continuous-native-generation)) or you're using native ios project manually, then you need to add the `NSLocationAlwaysAndWhenInUseUsageDescription`, `NSLocationAlwaysUsageDescription` and `NSLocationWhenInUseUsageDescription` keys to your project's ios/\[app\]/Info.plist:

```
<key>NSLocationAlwaysAndWhenInUseUsageDescription</key>
<string>Allow $(PRODUCT_NAME) to use your location</string>
<key>NSLocationAlwaysUsageDescription</key>
<string>Allow $(PRODUCT_NAME) to use your location</string>
<key>NSLocationWhenInUseUsageDescription</key>
<string>Allow $(PRODUCT_NAME) to use your location</string>
```

### Background location

Background location allows your app to receive location updates while it is running in the background and includes both location updates and region monitoring through geofencing. This feature is subject to platform API limitations and system constraints:

- Background location will stop if the user terminates the app.
- Background location resumes if the user restarts the app.
- Android
A terminated app will not automatically restart when a location
or geofencing event occurs due to platform limitations.
- iOS
The system will restart the terminated app when a new geofence
event occurs.

> On Android, the result of removing an app from the recent apps list varies by device vendor. For example, some implementations treat removing an app from the recent apps list as killing it. Read more about these differences here: [https://dontkillmyapp.com](https://dontkillmyapp.com/).

### Background location configuration  iOS

To be able to run background location on iOS, you need to add the `location` value to the `UIBackgroundModes` array in your app's Info.plist file.

If you're using [CNG](https://docs.expo.dev/workflow/continuous-native-generation), the required `UIBackgroundModes` configuration will be applied automatically by prebuild.

Configure UIBackgroundModes manually on iOS [Permalink](https://docs.expo.dev/versions/latest/sdk/location#configure-uibackgroundmodes-manually-on-ios)

If you're not using Continuous Native Generation ( [CNG](https://docs.expo.dev/workflow/continuous-native-generation)) or you're using a native iOS project, then you'll need to add the following to your Expo.plist file:

ios/project-name/Supporting/Expo.plist

Copy

```
<key>UIBackgroundModes</key>
  <array>
    <string>location</string>
  </array>
```

### Background location methods

To use Background Location methods, the following requirements apply:

- Location permissions must be granted.
- Background location task must be defined in the top-level scope, using [`TaskManager.defineTask`](https://docs.expo.dev/versions/latest/sdk/task-manager#taskmanagerdefinetasktaskname-taskexecutor).
- iOS
`"location"` background mode must be
specified in Info.plist file. See [Background location\\
configuration](https://docs.expo.dev/versions/latest/sdk/location#background-location-configuration).
- iOS
You must use a [development\\
build](https://docs.expo.dev/develop/development-builds/introduction) to use background location since it is not
supported in the Expo Go app.

### Geofencing methods

To use Geofencing methods, the following requirements apply:

- Location permissions must be granted.
- The Geofencing task must be defined in the top-level scope, using [`TaskManager.defineTask`](https://docs.expo.dev/versions/latest/sdk/task-manager#taskmanagerdefinetasktaskname-taskexecutor).

When using Geofencing, the following platform differences apply:

- Android
You are allowed [up to\\
100](https://developer.android.com/develop/sensors-and-location/location/geofencing) active
geofences per app.
- iOS
Expo Location will report the initial state of the registered
geofence(s) at app startup.
- iOS
There is a [limit of\\
20](https://developer.apple.com/documentation/corelocation/monitoring_the_user_s_proximity_to_geographic_regions)`regions` that can be simultaneously monitored.

### Background permissions

To use location tracking or Geofencing in the background, you must request the appropriate permissions:

- On Android, you must request both foreground and background permissions.
- On iOS, it must be granted with the `Always` option using [`requestBackgroundPermissionsAsync`](https://docs.expo.dev/versions/latest/sdk/location#locationrequestbackgroundpermissionsasync).

Expo and iOS permissions [Permalink](https://docs.expo.dev/versions/latest/sdk/location#expo-and-ios-permissions)

iOS permissions are divided into the two categories `When In Use` and `Always` and maps to Expo's foreground and background location permissions requested via:

- [`requestForegroundPermissionsAsync`](https://docs.expo.dev/versions/latest/sdk/location#locationrequestforegroundpermissionsasync) maps to `When In Use`
- [`requestBackgroundPermissionsAsync`](https://docs.expo.dev/versions/latest/sdk/location#locationrequestbackgroundpermissionsasync) maps to `Always`

> Note: When requesting `When In Use` authorization, the user can grant temporary access by selecting `Allow Once` in the system permission dialog. This authorization will be valid only for the current app session and is automatically revoked when the app is closed.

Detecting "Allow Once" versus "Allow While Using the App"

Unfortunately, iOS does not provide a way to detect whether the user selected `Allow Once` or `Allow While Using the App`. Both responses result in `When In Use` authorization.

If the user selected `Allow Once` and you subsequently call [`requestBackgroundPermissionsAsync`](https://docs.expo.dev/versions/latest/sdk/location#locationrequestbackgroundpermissionsasync) in the same session, the system will not show another prompt. Instead, the request will silently fail, and the returned background permission status will be denied.

Handling "Allow Once" scenarios

If you suspect the user selected `Allow Once` and needs to request background permissions, they must manually enable background location in the Settings app. You can use `Linking` to open the Settings app within your app:

```
import { Linking } from 'react-native';

function openSettings() {
  Linking.openURL('app-settings:');
}
```

Incremental permission requests

It is possible to request foreground location access first and then ask for background location access later. This can improve the user experience by requesting permissions only when necessary.

Requesting Background Permissions directly

If you call [`requestBackgroundPermissionsAsync`](https://docs.expo.dev/versions/latest/sdk/location#locationrequestbackgroundpermissionsasync) without first requesting foreground permissions, iOS treats it as a request for both `When In Use` and `Always` authorization. The system will then prompt the user for `When In Use` access, and the `Always` authorization prompt will be displayed when the system determines that `Always` authorization is required.

Remember that the user has the option of granting your app `When In Use` authorization instead. You must always be prepared to run with `When In Use` permission.

## Deferred locations

When using background locations, you can configure the location manager to defer updates. This helps save battery by reducing update frequency. You can set updates to trigger only after the device has moved a certain distance or after a specified time interval.

Deferred updates are configured through [`LocationTaskOptions`](https://docs.expo.dev/versions/latest/sdk/location#locationtaskoptions) using the [`deferredUpdatesDistance`](https://docs.expo.dev/versions/latest/sdk/location#locationtaskoptions), [`deferredUpdatesInterval`](https://docs.expo.dev/versions/latest/sdk/location#locationtaskoptions) and [`deferredTimeout`](https://docs.expo.dev/versions/latest/sdk/location#locationtaskoptions) properties.

> Deferred locations apply only when the app is in the background.

## Usage

If you're using the Android Emulator or iOS Simulator, ensure that [Location is enabled](https://docs.expo.dev/versions/latest/sdk/location#enable-emulator-location).

Location

Copy

Open in Snack

```
import { useState, useEffect } from 'react';
import { Platform, Text, View, StyleSheet } from 'react-native';
%%placeholder-start%%%%placeholder-end%%import * as Device from 'expo-device';
import * as Location from 'expo-location';

export default function App() {
  const [location, setLocation] = useState<Location.LocationObject | null>(null);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  useEffect(() => {
    async function getCurrentLocation() {
      %%placeholder-start%%%%placeholder-end%%if (Platform.OS === 'android' && !Device.isDevice) {
        setErrorMsg(
          'Oops, this will not work on Snack in an Android Emulator. Try it on your device!'
        );
        return;
      }
      let { status } = await Location.requestForegroundPermissionsAsync();
      if (status !== 'granted') {
        setErrorMsg('Permission to access location was denied');
        return;
      }

      let location = await Location.getCurrentPositionAsync({});
      setLocation(location);
    }

    getCurrentLocation();
  }, []);

  let text = 'Waiting...';
  if (errorMsg) {
    text = errorMsg;
  } else if (location) {
    text = JSON.stringify(location);
  }

  return (
    <View style={styles.container}>
      <Text style={styles.paragraph}>{text}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    padding: 20,
  },
  paragraph: {
    fontSize: 18,
    textAlign: 'center',
  },
});

Show more
```

## Enable emulator location

### Android Emulator

Open Android Studio, and launch the Android Emulator. Inside it, go to Settings \> Location and enable Use location.

![Location settings in Android Emulator for versions 12 and higher](https://docs.expo.dev/static/images/sdk/location/enable-android-emulator-location.png)

If you don't receive the locations in the emulator, you may have to turn off the Improve Location Accuracy setting. This will turn off Wi-Fi location and only use GPS. Then you can manipulate the location with GPS data through the emulator.

For Android 12 and higher, go to Settings \> Location \> Location Services \> Google Location Accuracy, and turn off Improve Location Accuracy. For Android 11 and lower, go to Settings \> Location \> Advanced \> Google Location Accuracy, and turn off Google Location Accuracy.

### iOS Simulator

With Simulator open, go to Features \> Location and choose any option besides None.

![Location settings in iOS simulator.](https://docs.expo.dev/static/images/sdk/location/ios-simulator-location.webp)

## API

```
import * as Location from 'expo-location';
```

## Hooks

### `useBackgroundPermissions(options)`

Android

iOS

Web

| Parameter | Type |
| --- | --- |
| options(optional) | `PermissionHookOptions<object>` |

Check or request permissions for the background location.
This uses both `requestBackgroundPermissionsAsync` and `getBackgroundPermissionsAsync` to
interact with the permissions.

Returns:

`[PermissionResponse | null, RequestPermissionMethod<PermissionResponse>, GetPermissionMethod<PermissionResponse>]`

Example

```
const [status, requestPermission] = Location.useBackgroundPermissions();
```

### `useForegroundPermissions(options)`

Android

iOS

Web

| Parameter | Type |
| --- | --- |
| options(optional) | `PermissionHookOptions<object>` |

Check or request permissions for the foreground location.
This uses both `requestForegroundPermissionsAsync` and `getForegroundPermissionsAsync` to interact with the permissions.

Returns:

`[LocationPermissionResponse | null, RequestPermissionMethod<LocationPermissionResponse>, GetPermissionMethod<LocationPermissionResponse>]`

Example

```
const [status, requestPermission] = Location.useForegroundPermissions();
```

### `useMotionActivityPermissions(options)`

Android

iOS

| Parameter | Type |
| --- | --- |
| options(optional) | `PermissionHookOptions<object>` |

Checks or requests permissions for motion activity detection.
This uses both `requestMotionActivityPermissionsAsync` and `getMotionActivityPermissionsAsync`
to interact with the permissions.

Returns:

`[PermissionResponse | null, RequestPermissionMethod<PermissionResponse>, GetPermissionMethod<PermissionResponse>]`

Example

```
const [status, requestPermission] = Location.useMotionActivityPermissions();
```

## Methods

### `Location.enableNetworkProviderAsync()`

Android

Asks the user to turn on high accuracy location mode which enables network provider that uses
Google Play services to improve location accuracy and location-based services.

Returns:

`Promise<void>`

A promise resolving as soon as the user accepts the dialog. Rejects if denied.

### `Location.geocodeAsync(address)`

Android

iOS

| Parameter | Type | Description |
| --- | --- | --- |
| address | `string` | A string representing address, eg. `"Baker Street London"`. |

Geocode an address string to latitude-longitude location.

On Android, you must request location permissions with [`requestForegroundPermissionsAsync`](https://docs.expo.dev/versions/latest/sdk/location#locationrequestforegroundpermissionsasync)
before geocoding can be used.

> Note: Geocoding is resource consuming and has to be used reasonably. Creating too many
> requests at a time can result in an error, so they have to be managed properly.
> It's also discouraged to use geocoding while the app is in the background and its results won't
> be shown to the user immediately.

Returns:

`Promise<LocationGeocodedLocation[]>`

A promise which fulfills with an array (in most cases its size is 1) of [`LocationGeocodedLocation`](https://docs.expo.dev/versions/latest/sdk/location#locationgeocodedlocation)
objects.

### `Location.getBackgroundPermissionsAsync()`

Android

iOS

Web

Checks user's permissions for accessing location while the app is in the background.

Returns:

`Promise<PermissionResponse>`

A promise that fulfills with an object of type [`PermissionResponse`](https://docs.expo.dev/versions/latest/sdk/location#permissionresponse).

### `Location.getCurrentPositionAsync(options)`

Android

iOS

Web

| Parameter | Type |
| --- | --- |
| options(optional) | `LocationOptions` |

Requests for one-time delivery of the user's current location.
Depending on given `accuracy` option it may take some time to resolve,
especially when you're inside a building.

> Note: Calling it causes the location manager to obtain a location fix which may take several
> seconds. Consider using [`getLastKnownPositionAsync`](https://docs.expo.dev/versions/latest/sdk/location#locationgetlastknownpositionasyncoptions)
> if you expect to get a quick response and high accuracy is not required.

Returns:

`Promise<LocationObject>`

A promise which fulfills with an object of type [`LocationObject`](https://docs.expo.dev/versions/latest/sdk/location#locationobject).

### `Location.getForegroundPermissionsAsync()`

Android

iOS

Web

Checks user's permissions for accessing location while the app is in the foreground.

Returns:

`Promise<LocationPermissionResponse>`

A promise that fulfills with an object of type [`LocationPermissionResponse`](https://docs.expo.dev/versions/latest/sdk/location#locationpermissionresponse).

### `Location.getHeadingAsync()`

Android

iOS

Web

Gets the current heading information from the device. To simplify, it calls `watchHeadingAsync`
and waits for a couple of updates, and then returns the one that is accurate enough.

Returns:

`Promise<LocationHeadingObject>`

A promise which fulfills with an object of type [`LocationHeadingObject`](https://docs.expo.dev/versions/latest/sdk/location#locationheadingobject).

### `Location.getLastKnownPositionAsync(options)`

Android

iOS

Web

| Parameter | Type |
| --- | --- |
| options(optional) | `LocationLastKnownOptions` |

Gets the last known position of the device or `null` if it's not available or doesn't match given
requirements such as maximum age or required accuracy.
It's considered to be faster than `getCurrentPositionAsync` as it doesn't request for the current
location, but keep in mind the returned location may not be up-to-date.

Returns:

`Promise<LocationObject | null>`

A promise which fulfills with an object of type [`LocationObject`](https://docs.expo.dev/versions/latest/sdk/location#locationobject) or
`null` if it's not available or doesn't match given requirements such as maximum age or required
accuracy.

### `Location.getMotionActivityAsync()`

Android

iOS

Fetches the current motion activity status of the device by subscribing to the first available
activity update and immediately unsubscribing afterwards.
The returned promise fulfills with the same object shape as the `watchMotionActivityAsync`
callback.

The method uses the Google Play Services activity
recognition (Android) or platform's motion coprocessor (iOS) and does not require location permissions.
On Android 10+, the `ACTIVITY_RECOGNITION` runtime permission must be granted beforehand.
On iOS, the system will prompt the user for Motion and Fitness access the first time this method is called.

Returns:

`Promise<MotionActivityObject>`

a promise which fulfills with a [`MotionActivityObject`](https://docs.expo.dev/versions/latest/sdk/location#motionactivityobject).

### `Location.getMotionActivityPermissionsAsync()`

Android

iOS

Checks user's permissions for accessing motion activity data.

Returns:

`Promise<PermissionResponse>`

A promise that fulfills with an object of type [`PermissionResponse`](https://docs.expo.dev/versions/latest/sdk/location#permissionresponse).

### `Location.getProviderStatusAsync()`

Android

iOS

Web

Check status of location providers.

Returns:

`Promise<LocationProviderStatus>`

A promise which fulfills with an object of type [`LocationProviderStatus`](https://docs.expo.dev/versions/latest/sdk/location#locationproviderstatus).

### `Location.hasServicesEnabledAsync()`

Android

iOS

Web

Checks whether location services are enabled by the user.

Returns:

`Promise<boolean>`

A promise which fulfills to `true` if location services are enabled on the device,
or `false` if not.

### `Location.hasStartedGeofencingAsync(taskName)`

Android

iOS

Web

| Parameter | Type | Description |
| --- | --- | --- |
| taskName | `string` | Name of the geofencing task to check. |

Returns:

`Promise<boolean>`

A promise which fulfills with boolean value indicating whether the geofencing task is
started or not.

### `Location.hasStartedLocationUpdatesAsync(taskName)`

Android

iOS

Web

| Parameter | Type | Description |
| --- | --- | --- |
| taskName | `string` | Name of the location task to check. |

Returns:

`Promise<boolean>`

A promise which fulfills with boolean value indicating whether the location task is
started or not.

### `Location.installWebGeolocationPolyfill()`

Android

iOS

Web

Polyfills `navigator.geolocation` for interop with the core React Native and Web API approach to geolocation.

Returns:

`void`

### `Location.isBackgroundLocationAvailableAsync()`

Android

iOS

Web

Returns:

`Promise<boolean>`

### `Location.requestBackgroundPermissionsAsync()`

Android

iOS

Web

Asks the user to grant permissions for location while the app is in the background.
On Android 11 or higher: this method will open the system settings page - before that happens
you should explain to the user why your application needs background location permission.
For example, you can use `Modal` component from `react-native` to do that.

> Note: Foreground permissions should be granted before asking for the background permissions
> (your app can't obtain background permission without foreground permission).

Returns:

`Promise<PermissionResponse>`

A promise that fulfills with an object of type [`PermissionResponse`](https://docs.expo.dev/versions/latest/sdk/location#permissionresponse).

### `Location.requestForegroundPermissionsAsync()`

Android

iOS

Web

Asks the user to grant permissions for location while the app is in the foreground.

Returns:

`Promise<LocationPermissionResponse>`

A promise that fulfills with an object of type [`LocationPermissionResponse`](https://docs.expo.dev/versions/latest/sdk/location#locationpermissionresponse).

### `Location.requestMotionActivityPermissionsAsync()`

Android

iOS

Asks the user to grant permissions for motion activity detection.
On Android 10+, this requests the `ACTIVITY_RECOGNITION` runtime permission.
On iOS, this triggers the system prompt for Motion and Fitness access the first time it is called.

Returns:

`Promise<PermissionResponse>`

a promise that fulfills with an object of type [`PermissionResponse`](https://docs.expo.dev/versions/latest/sdk/location#permissionresponse).

### `Location.reverseGeocodeAsync(location)`

Android

iOS

| Parameter | Type | Description |
| --- | --- | --- |
| location | `Pick<LocationGeocodedLocation, 'latitude' | 'longitude'>` | An object representing a location. |

Reverse geocode a location to postal address.

On Android, you must request location permissions with [`requestForegroundPermissionsAsync`](https://docs.expo.dev/versions/latest/sdk/location#locationrequestforegroundpermissionsasync)
before geocoding can be used.

> Note: Geocoding is resource consuming and has to be used reasonably. Creating too many
> requests at a time can result in an error, so they have to be managed properly.
> It's also discouraged to use geocoding while the app is in the background and its results won't
> be shown to the user immediately.

Returns:

`Promise<LocationGeocodedAddress[]>`

A promise which fulfills with an array (in most cases its size is 1) of [`LocationGeocodedAddress`](https://docs.expo.dev/versions/latest/sdk/location#locationgeocodedaddress) objects.

### `Location.startGeofencingAsync(taskName, regions)`

Android

iOS

Web

| Parameter | Type | Description |
| --- | --- | --- |
| taskName | `string` | Name of the task that will be called when the device enters or exits from specified regions. |
| regions(optional) | `LocationRegion[]` | Array of region objects to be geofenced.<br>Default:`[]` |

Starts geofencing for given regions. When the new event comes, the task with specified name will
be called with the region that the device enter to or exit from.
If you want to add or remove regions from already running geofencing task, you can just call
`startGeofencingAsync` again with the new array of regions.

#### Task parameters

Geofencing task will be receiving following data:

- `eventType` \- Indicates the reason for calling the task, which can be triggered by entering or exiting the region.
See [`GeofencingEventType`](https://docs.expo.dev/versions/latest/sdk/location#geofencingeventtype).
- `region` \- Object containing details about updated region. See [`LocationRegion`](https://docs.expo.dev/versions/latest/sdk/location#locationregion) for more details.

Returns:

`Promise<void>`

A promise resolving as soon as the task is registered.

Example

```
import { GeofencingEventType } from 'expo-location';
import * as TaskManager from 'expo-task-manager';

 TaskManager.defineTask(YOUR_TASK_NAME, ({ data: { eventType, region }, error }) => {
  if (error) {
    // check `error.message` for more details.
    return;
  }
  if (eventType === GeofencingEventType.Enter) {
    console.log("You've entered region:", region);
  } else if (eventType === GeofencingEventType.Exit) {
    console.log("You've left region:", region);
  }
});
```

### `Location.startLocationUpdatesAsync(taskName, options)`

Android

iOS

Web

| Parameter | Type | Description |
| --- | --- | --- |
| taskName | `string` | Name of the task receiving location updates. |
| options(optional) | `LocationTaskOptions` | An object of options passed to the location manager. |

Registers for receiving location updates that can also come when the app is in the background.

#### Task parameters

Background location task will be receiving following data:

- `locations` \- An array of the new locations.

Returns:

`Promise<void>`

A promise resolving once the task with location updates is registered.

Example

```
import * as TaskManager from 'expo-task-manager';

TaskManager.defineTask(YOUR_TASK_NAME, ({ data: { locations }, error }) => {
 if (error) {
   // check `error.message` for more details.
   return;
 }
 console.log('Received new locations', locations);
});
```

### `Location.stopGeofencingAsync(taskName)`

Android

iOS

Web

| Parameter | Type | Description |
| --- | --- | --- |
| taskName | `string` | Name of the task to unregister. |

Stops geofencing for specified task. It unregisters the background task so the app will not be
receiving any updates, especially in the background.

Returns:

`Promise<void>`

A promise resolving as soon as the task is unregistered.

### `Location.stopLocationUpdatesAsync(taskName)`

Android

iOS

Web

| Parameter | Type | Description |
| --- | --- | --- |
| taskName | `string` | Name of the background location task to stop. |

Stops location updates for specified task.

Returns:

`Promise<void>`

A promise resolving as soon as the task is unregistered.

### `Location.watchHeadingAsync(callback, errorHandler)`

Android

iOS

| Parameter | Type | Description |
| --- | --- | --- |
| callback | `LocationHeadingCallback` | This function is called on each compass update. It receives an object of type<br>[LocationHeadingObject](https://docs.expo.dev/versions/latest/sdk/location#locationheadingobject) as the first argument. |
| errorHandler(optional) | `LocationErrorCallback` | This function is called when an error occurs. It receives a string with the<br>error message as the first argument. |

Subscribe to compass updates from the device.

Returns:

`Promise<LocationSubscription>`

A promise which fulfills with a [`LocationSubscription`](https://docs.expo.dev/versions/latest/sdk/location#locationsubscription) object.

### `Location.watchMotionActivityAsync(callback, errorHandler)`

Android

iOS

| Parameter | Type | Description |
| --- | --- | --- |
| callback | `MotionActivityCallback` | This function is called on each motion activity update. It receives an object<br>of type [`MotionActivityObject`](https://docs.expo.dev/versions/latest/sdk/location#motionactivityobject) as the first argument. |
| errorHandler(optional) | `LocationErrorCallback` | This function is called if the platform reports an error (for example,<br>when activity recognition permission is denied). It receives a string message as the<br>first argument. |

Subscribes to motion activity updates from the device. The callback fires whenever the
platform's motion coprocessor detects a change in the user's activity. Only foreground use
is supported - updates pause when the app is backgrounded and resume when it returns to
the foreground.

Returns:

`Promise<LocationSubscription>`

a promise which fulfills with a [`LocationSubscription`](https://docs.expo.dev/versions/latest/sdk/location#locationsubscription) object.

### `Location.watchPositionAsync(options, callback, errorHandler)`

Android

iOS

Web

| Parameter | Type | Description |
| --- | --- | --- |
| options | `LocationOptions` | - |
| callback | `LocationCallback` | This function is called on each location update. It receives an object of type<br>[`LocationObject`](https://docs.expo.dev/versions/latest/sdk/location#locationobject) as the first argument. |
| errorHandler(optional) | `LocationErrorCallback` | This function is called when an error occurs. It receives a string with the<br>error message as the first argument. |

Subscribe to location updates from the device. Updates will only occur while the application is in
the foreground. To get location updates while in background you'll need to use
[`startLocationUpdatesAsync`](https://docs.expo.dev/versions/latest/sdk/location#locationstartlocationupdatesasynctaskname-options).

Returns:

`Promise<LocationSubscription>`

A promise which fulfills with a [`LocationSubscription`](https://docs.expo.dev/versions/latest/sdk/location#locationsubscription) object.

## Types

### `LocationCallback(location)`

Android

iOS

Web

Represents `watchPositionAsync` callback.

| Parameter | Type |
| --- | --- |
| location | `LocationObject` |

Returns:

`any`

### `LocationErrorCallback(reason)`

Android

iOS

Web

Error callback for location methods.

| Parameter | Type |
| --- | --- |
| reason | `string` |

Returns:

`void`

### `LocationGeocodedAddress`

Android

iOS

Web

Type representing a result of `reverseGeocodeAsync`.

| Property | Type | Description |
| --- | --- | --- |
| city | `string | null` | City name of the address. |
| country | `string | null` | Localized country name of the address. |
| district | `string | null` | Additional city-level information like district name. |
| formattedAddress | `string | null` | Only for:<br>Android<br>Composed string of the address components, for example, "111 8th Avenue, New York, NY". |
| isoCountryCode | `string | null` | Localized (ISO) country code of the address, if available. |
| name | `string | null` | The name of the placemark, for example, "Tower Bridge". |
| postalCode | `string | null` | Postal code of the address. |
| region | `string | null` | The state or province associated with the address. |
| street | `string | null` | Street name of the address. |
| streetNumber | `string | null` | Street number of the address. |
| subregion | `string | null` | Additional information about administrative area. |
| timezone | `string | null` | Only for:<br>iOS<br>The timezone identifier associated with the address. |

### `LocationGeocodedLocation`

Android

iOS

Web

Type representing a result of `geocodeAsync`.

| Property | Type | Description |
| --- | --- | --- |
| accuracy(optional) | `number` | The radius of uncertainty for the location, measured in meters. |
| altitude(optional) | `number` | The altitude in meters above the WGS 84 reference ellipsoid. |
| latitude | `number` | The latitude in degrees. |
| longitude | `number` | The longitude in degrees. |

### `LocationHeadingCallback(location)`

Android

iOS

Web

Represents `watchHeadingAsync` callback.

| Parameter | Type |
| --- | --- |
| location | `LocationHeadingObject` |

Returns:

`any`

### `LocationHeadingObject`

Android

iOS

Web

Type of the object containing heading details and provided by `watchHeadingAsync` callback.

| Property | Type | Description |
| --- | --- | --- |
| accuracy | `number` | Level of calibration of compass:<br>- `3`: high accuracy<br>- `2`: medium accuracy<br>- `1`: low accuracy<br>- `0`: none<br>Reference for iOS:<br>- `3`: < 20 degrees uncertainty<br>- `2`: < 35 degrees<br>- `1`: < 50 degrees<br>- `0`: \> 50 degrees |
| magHeading | `number` | Measure of magnetic north in degrees. |
| trueHeading | `number` | Measure of true north in degrees (needs location permissions, will return `-1` if not given). |

### `LocationLastKnownOptions`

Android

iOS

Web

Type representing options object that can be passed to `getLastKnownPositionAsync`.

| Property | Type | Description |
| --- | --- | --- |
| maxAge(optional) | `number` | A number of milliseconds after which the last known location starts to be invalid and thus<br>`null` is returned. |
| requiredAccuracy(optional) | `number` | The maximum radius of uncertainty for the location, measured in meters. If the last known<br>location's accuracy radius is bigger (less accurate) then `null` is returned. |

### `LocationObject`

Android

iOS

Web

Type representing the location object.

| Property | Type | Description |
| --- | --- | --- |
| coords | `LocationObjectCoords` | The coordinates of the position. |
| mocked(optional) | `boolean` | Only for:<br>Android<br>Whether the location coordinates is mocked or not. |
| timestamp | `number` | The time at which this position information was obtained, in milliseconds since epoch. |

### `LocationObjectCoords`

Android

iOS

Web

Type representing the location GPS related data.

| Property | Type | Description |
| --- | --- | --- |
| accuracy | `number | null` | The radius of uncertainty for the location, measured in meters. Can be `null` on Web if it's not available. |
| altitude | `number | null` | The altitude in meters above the WGS 84 reference ellipsoid. Can be `null` on Web if it's not available. |
| altitudeAccuracy | `number | null` | The accuracy of the altitude value, in meters. Can be `null` on Web if it's not available. |
| heading | `number | null` | Horizontal direction of travel of this device, measured in degrees starting at due north and<br>continuing clockwise around the compass. Thus, north is 0 degrees, east is 90 degrees, south is<br>180 degrees, and so on. Can be `null` on Web if it's not available. |
| latitude | `number` | The latitude in degrees. |
| longitude | `number` | The longitude in degrees. |
| speed | `number | null` | The instantaneous speed of the device in meters per second. Can be `null` on Web if it's not available. |

### `LocationOptions`

Android

iOS

Web

Type representing options argument in `getCurrentPositionAsync`.

| Property | Type | Description |
| --- | --- | --- |
| accuracy(optional) | `Accuracy` | Location manager accuracy. Pass one of `Accuracy` enum values.<br>For low-accuracies the implementation can avoid geolocation providers<br>that consume a significant amount of power (such as GPS).<br>Default:`LocationAccuracy.Balanced` |
| distanceInterval(optional) | `number` | Receive updates only when the location has changed by at least this distance in meters.<br>Default value may depend on `accuracy` option. |
| mayShowUserSettingsDialog(optional) | `boolean` | Only for:<br>Android<br>Specifies whether to ask the user to turn on improved accuracy location mode<br>which uses Wi-Fi, cell networks and GPS sensor.<br>Default:`true` |
| timeInterval(optional) | `number` | Only for:<br>Android<br>Minimum time to wait between each update in milliseconds.<br>Default value may depend on `accuracy` option. |

### `LocationPermissionResponse`

Android

iOS

Web

`LocationPermissionResponse` extends [`PermissionResponse`](https://docs.expo.dev/versions/latest/sdk/location#permissionresponse)
type exported by `expo-modules-core` and contains additional platform-specific fields.

Type:`PermissionResponse`extended by:

| Property | Type | Description |
| --- | --- | --- |
| android(optional) | `PermissionDetailsLocationAndroid` | - |
| ios(optional) | `PermissionDetailsLocationIOS` | - |

### `LocationProviderStatus`

Android

iOS

Web

Represents the object containing details about location provider.

| Property | Type | Description |
| --- | --- | --- |
| backgroundModeEnabled | `boolean` | - |
| gpsAvailable(optional) | `boolean` | Only for:<br>Android<br>Whether the GPS provider is available. If `true` the location data will come<br>from GPS, especially for requests with high accuracy. |
| locationServicesEnabled | `boolean` | Whether location services are enabled. See [Location.hasServicesEnabledAsync](https://docs.expo.dev/versions/latest/sdk/location#locationhasservicesenabledasync)<br>for a more convenient solution to get this value. |
| networkAvailable(optional) | `boolean` | Only for:<br>Android<br>Whether the network provider is available. If `true` the location data will<br>come from cellular network, especially for requests with low accuracy. |
| passiveAvailable(optional) | `boolean` | Only for:<br>Android<br>Whether the passive provider is available. If `true` the location data will<br>be determined passively. |

### `LocationRegion`

Android

iOS

Web

Type representing geofencing region object.

| Property | Type | Description |
| --- | --- | --- |
| identifier(optional) | `string` | The identifier of the region object. Defaults to auto-generated UUID hash. |
| latitude | `number` | The latitude in degrees of region's center point. |
| longitude | `number` | The longitude in degrees of region's center point. |
| notifyOnEnter(optional) | `boolean` | Boolean value whether to call the task if the device enters the region.<br>Default:`true` |
| notifyOnExit(optional) | `boolean` | Boolean value whether to call the task if the device exits the region.<br>Default:`true` |
| radius | `number` | The radius measured in meters that defines the region's outer boundary. |
| state(optional) | `GeofencingRegionState` | One of [GeofencingRegionState](https://docs.expo.dev/versions/latest/sdk/location#geofencingregionstate) region state. Determines whether the<br>device is inside or outside a region. |

### `LocationSubscription`

Android

iOS

Web

Represents subscription object returned by methods watching for new locations or headings.

| Property | Type | Description |
| --- | --- | --- |
| remove | `() => void` | Call this function with no arguments to remove this subscription. The callback will no longer<br>be called for location updates. |

### `LocationTaskOptions`

Android

iOS

Web

Type representing background location task options.

Type:`LocationOptions`extended by:

| Property | Type | Description |
| --- | --- | --- |
| activityType(optional) | `ActivityType` | Only for:<br>iOS<br>The type of user activity associated with the location updates.<br>Default:`ActivityType.Other`<br>> See: See [Apple docs](https://developer.apple.com/documentation/corelocation/cllocationmanager/1620567-activitytype) for more details. |
| deferredUpdatesDistance(optional) | `number` | The distance in meters that must occur between last reported location and the current location<br>before deferred locations are reported.<br>Default:`0` |
| deferredUpdatesInterval(optional) | `number` | Minimum time interval in milliseconds that must pass since last reported location before all<br>later locations are reported in a batched update<br>Default:`0` |
| deferredUpdatesTimeout(optional) | `number` | - |
| foregroundService(optional) | `LocationTaskServiceOptions` | - |
| pausesUpdatesAutomatically(optional) | `boolean` | Only for:<br>iOS<br>A boolean value indicating whether the location manager can pause location<br>updates to improve battery life without sacrificing location data. When this option is set to<br>`true`, the location manager pauses updates (and powers down the appropriate hardware) at times<br>when the location data is unlikely to change. You can help the determination of when to pause<br>location updates by assigning a value to the `activityType` property.<br>Default:`false` |
| showsBackgroundLocationIndicator(optional) | `boolean` | Only for:<br>iOS<br>A boolean indicating whether the status bar changes its appearance when<br>location services are used in the background.<br>Default:`false` |

### `LocationTaskServiceOptions`

Android

iOS

Web

| Property | Type | Description |
| --- | --- | --- |
| killServiceOnDestroy(optional) | `boolean` | Boolean value whether to destroy the foreground service if the app is killed. |
| notificationBody | `string` | Subtitle of the foreground service notification. |
| notificationColor(optional) | `string` | Color of the foreground service notification. Accepts `#RRGGBB` and `#AARRGGBB` hex formats. |
| notificationTitle | `string` | Title of the foreground service notification. |

### `MotionActivityCallback(activity)`

Android

iOS

Web

Represents the `watchMotionActivityAsync` callback.

| Parameter | Type |
| --- | --- |
| activity | `MotionActivityObject` |

Returns:

`any`

### `MotionActivityObject`

Android

iOS

Web

Type returned by `getMotionActivityAsync` and the `watchMotionActivityAsync` callback.
Contains one entry per `MotionActivityType` so callers can inspect each activity
independently without searching through an array.

Example

```
const { activities } = await Location.getMotionActivityAsync();
if (activities.automotive.detected) {
  console.log('driving, confidence:', activities.automotive.confidence);
}
```

| Property | Type | Description |
| --- | --- | --- |
| activities | `Record<MotionActivityType, MotionActivityState>` | An object keyed by `MotionActivityType`, each value being a `MotionActivityState`. |
| timestamp | `number` | The time at which the activity snapshot was taken, in milliseconds since Unix epoch. |

### `MotionActivityState`

Android

iOS

Web

Detection state for a single activity type.
When `detected` is `false`, `confidence` is always `Low`.

| Property | Type | Description |
| --- | --- | --- |
| confidence | `MotionActivityConfidence` | Confidence level of the detection.<br>On Android, each activity type has its own confidence derived from the<br>`DetectedActivity` probability value. On iOS, this is the overall reading confidence from<br>`CMMotionActivityConfidence`, shared by all detected activities. |
| detected | `boolean` | Whether this activity was detected. |

### `PermissionDetailsLocationAndroid`

Android

iOS

Web

| Property | Type | Description |
| --- | --- | --- |
| accuracy | `'fine' | 'coarse' | 'none'` | Indicates the type of location provider. |

### `PermissionDetailsLocationIOS`

Android

iOS

Web

| Property | Type | Description |
| --- | --- | --- |
| accuracy | `'full' | 'reduced'` | Only for:<br>iOS 14+<br>The accuracy authorization granted by the user.<br>`'full'` means the user has granted precise location access, `'reduced'` means the app<br>can only access approximate location. Below iOS 14 always returns `'full'`. |
| scope | `'whenInUse' | 'always' | 'none'` | The scope of granted permission. Indicates when it's possible to use location. |

### `PermissionExpiration`

Android

iOS

Web

Literal type: `union`

Permission expiration time. Currently, all permissions are granted permanently.

Acceptable values are:`'never'` \| `number`

### `PermissionHookOptions`

Android

iOS

Web

Literal type: `union`

Acceptable values are:`PermissionHookBehavior` \| `Options`

### `PermissionResponse`

Android

iOS

Web

An object obtained by permissions get and request functions.

| Property | Type | Description |
| --- | --- | --- |
| canAskAgain | `boolean` | Indicates if user can be asked again for specific permission.<br>If not, one should be directed to the Settings app<br>in order to enable/disable the permission. |
| expires | `PermissionExpiration` | Determines time when the permission expires. |
| granted | `boolean` | A convenience boolean that indicates if the permission is granted. |
| status | `PermissionStatus` | Determines the status of the permission. |

## Enums

### `Accuracy`

Android

iOS

Web

Enum with available location accuracies.

#### `Lowest`

`Accuracy.Lowest ＝ 1`

Accurate to the nearest three kilometers.

#### `Low`

`Accuracy.Low ＝ 2`

Accurate to the nearest kilometer.

#### `Balanced`

`Accuracy.Balanced ＝ 3`

Accurate to within one hundred meters.

#### `High`

`Accuracy.High ＝ 4`

Accurate to within ten meters of the desired target.

#### `Highest`

`Accuracy.Highest ＝ 5`

The best level of accuracy available.

#### `BestForNavigation`

`Accuracy.BestForNavigation ＝ 6`

The highest possible accuracy that uses additional sensor data to facilitate navigation apps.

### `ActivityType`

Android

iOS

Web

Enum with available activity types of background location tracking.

#### `Other`

`ActivityType.Other ＝ 1`

Default activity type. Use it if there is no other type that matches the activity you track.

#### `AutomotiveNavigation`

`ActivityType.AutomotiveNavigation ＝ 2`

Location updates are being used specifically during vehicular navigation to track location
changes to the automobile.

#### `Fitness`

`ActivityType.Fitness ＝ 3`

Use this activity type if you track fitness activities such as walking, running, cycling,
and so on.

#### `OtherNavigation`

`ActivityType.OtherNavigation ＝ 4`

Activity type for movements for other types of vehicular navigation that are not automobile
related.

#### `Airborne`

iOS

`ActivityType.Airborne ＝ 5`

Intended for airborne activities. Fall backs to `ActivityType.Other` if
unsupported.

### `GeofencingEventType`

Android

iOS

Web

A type of the event that geofencing task can receive.

#### `Enter`

`GeofencingEventType.Enter ＝ 1`

Emitted when the device entered observed region.

#### `Exit`

`GeofencingEventType.Exit ＝ 2`

Occurs as soon as the device left observed region

### `GeofencingRegionState`

Android

iOS

Web

State of the geofencing region that you receive through the geofencing task.

#### `Unknown`

`GeofencingRegionState.Unknown ＝ 0`

Indicates that the device position related to the region is unknown.

#### `Inside`

`GeofencingRegionState.Inside ＝ 1`

Indicates that the device is inside the region.

#### `Outside`

`GeofencingRegionState.Outside ＝ 2`

Inverse of inside state.

### `MotionActivityConfidence`

Android

iOS

Web

Confidence level for motion activity detection. On Android, the raw `DetectedActivity` confidence (0-
100) is bucketed into these three levels. On iOS, maps directly to `CMMotionActivityConfidence`.

#### `Low`

`MotionActivityConfidence.Low ＝ 0`

The activity determination has low confidence.

#### `Medium`

`MotionActivityConfidence.Medium ＝ 1`

The activity determination has medium confidence.

#### `High`

`MotionActivityConfidence.High ＝ 2`

The activity determination has high confidence.

### `MotionActivityType`

Android

iOS

Web

The type of physical activity the user is currently performing.

On Android it maps to `DetectedActivity` constants from Google Play Services.
On iOS this maps to the boolean properties of `CMMotionActivity` (the highest-priority
truthy property wins).

#### `Automotive`

`MotionActivityType.Automotive ＝ "automotive"`

The device is in a motorized vehicle (car, bus, train, and so on).

#### `Cycling`

`MotionActivityType.Cycling ＝ "cycling"`

The user is riding a bicycle.

#### `Running`

`MotionActivityType.Running ＝ "running"`

The user is running.

#### `Stationary`

`MotionActivityType.Stationary ＝ "stationary"`

The device is not moving.

#### `Unknown`

`MotionActivityType.Unknown ＝ "unknown"`

The activity cannot be determined.

#### `Walking`

`MotionActivityType.Walking ＝ "walking"`

The user is walking.

### `PermissionStatus`

Android

iOS

Web

#### `DENIED`

`PermissionStatus.DENIED ＝ "denied"`

User has denied the permission.

#### `GRANTED`

`PermissionStatus.GRANTED ＝ "granted"`

User has granted the permission.

#### `UNDETERMINED`

`PermissionStatus.UNDETERMINED ＝ "undetermined"`

User hasn't granted or denied the permission yet.

## Permissions

### Android

> Foreground and background services are not available in Expo Go for Android. Instead, we recommend using a [development build](https://docs.expo.dev/develop/development-builds/introduction) to avoid limitations.

When you install the `expo-location` module, it automatically adds the following permissions:

- `ACCESS_COARSE_LOCATION`: for approximate device location
- `ACCESS_FINE_LOCATION`: for precise device location

The following permissions are optional:

- `FOREGROUND_SERVICE` and `FOREGROUND_SERVICE_LOCATION`: to be able to access location while the app is open but backgrounded. `FOREGROUND_SERVICE_LOCATION` is only required as of Android 14. When you enable this in a new build, you will need to [submit your app for review and request access to use the foreground service permission](https://support.google.com/googleplay/android-developer/answer/13392821?hl=en).
- `ACCESS_BACKGROUND_LOCATION`: to be able to access location while the app is backgrounded or closed. When you enable this in a new build, you will need to [submit your app for review and request access to use the background location permission](https://support.google.com/googleplay/android-developer/answer/9799150?hl=en).

| Android permission | Description |
| --- | --- |
| `ACCESS_COARSE_LOCATION` | Allows an app to access approximate location.<br>> Alternatively, you might want `ACCESS_FINE_LOCATION`. |
| `ACCESS_FINE_LOCATION` | Allows an app to access precise location.<br>> Alternatively, you might want `ACCESS_COARSE_LOCATION`. |
| `FOREGROUND_SERVICE` | Allows a regular application to use Service.startForeground.<br>> Allows a regular application to use `Service.startForeground`. |
| `FOREGROUND_SERVICE_LOCATION` | Allows a regular application to use Service.startForeground with the type "location".<br>> Allows a regular application to use `Service.startForeground` with the type "location". |
| `ACCESS_BACKGROUND_LOCATION` | Allows an app to access location in the background.<br>> If you're requesting this permission, you must also request either `ACCESS_COARSE_LOCATION` or `ACCESS_FINE_LOCATION`. Requesting this permission by itself doesn't give you location access. |

#### Excluding a permission

> Note: Excluding a required permission from a module in your app can break the functionality corresponding to that permission. Always make sure to include all permissions a module is dependent on.

When your Expo project doesn't benefit from having particular permission included, you can omit it. For example, if your application doesn't need access to the precise location, you can exclude the `ACCESS_FINE_LOCATION` permission.

Another example can be stated using [available location accuracies](https://docs.expo.dev/versions/latest/sdk/location#accuracy). Android defines the approximate location accuracy estimation within about 3 square kilometers, and the precise location accuracy estimation within about 50 meters. For example, if the location accuracy value is [Low](https://docs.expo.dev/versions/latest/sdk/location#low), you can exclude `ACCESS_FINE_LOCATION` permission. To learn more about levels of location accuracies, see [Android documentation](https://developer.android.com/training/location/permissions#accuracy).

To learn more on how to exclude permission, see [Excluding Android permissions](https://docs.expo.dev/guides/permissions#android).

### iOS

The following usage description keys are used by this library:

| Info.plist key | Description |
| --- | --- |
| `NSLocationAlwaysAndWhenInUseUsageDescription` | A message that tells the user why the app is requesting access to the user’s location information at all times. |
| `NSLocationAlwaysUsageDescription` | A message that tells the user why the app is requesting access to the user's location at all times.DeprecatedFor apps deployed to targets in iOS 11 and later, use NSLocationAlwaysAndWhenInUseUsageDescription instead. |
| `NSLocationWhenInUseUsageDescription` | A message that tells the user why the app is requesting access to the user’s location information while the app is running in the foreground. |

`NSLocationAlwaysUsageDescription` is deprecated in favor of `NSLocationAlwaysAndWhenInUseUsageDescription` from iOS 11.

We value your privacy

We use cookies to collect data and improve our services. [Learn more](https://expo.dev/privacy/cookies)

DeclineAccept

Customize
