# BrainBlitz for Android

Sign in or create a BrainBlitz account, view your profile and statistics, and search for a multiplayer match. The app saves the access token in app-private storage encrypted with a key held by Android Keystore. When the token expires, sign in again. [Android Keystore](https://developer.android.com/privacy-and-security/keystore)

## Run locally

1. Open `clients/android` in Android Studio. Use its bundled JDK and install Android SDK Platform 35 and Build Tools 35.0.0 through SDK Manager if needed. The app supports Android 8.0 (API 26) and newer.
2. Start the backend from the repository root with `docker compose up -d --wait`. The default debug build connects to the host gateway at `http://10.0.2.2:8088` on the Android emulator.
3. Run the `app` configuration on two emulators. Create a separate test account on each one or sign in with two existing accounts. Open **Play** on each device and choose the same category, timer, and total player count of two. Select **Find match** on both devices and check that both debug builds display the same match ID. [Backend setup](../../docs/run-on-another-computer.md)

## Build and test

From `clients/android`, with JDK 17 or newer and the Android SDK configured, run:

```sh
./gradlew :app:assembleDebug :app:testDebugUnitTest :app:lintDebug
```

The debug APK is at `app/build/outputs/apk/debug/app-debug.apk`. To build against a staging gateway, set `brainBlitzStagingUrl` to its HTTPS origin in your Gradle properties or pass `-PbrainBlitzStagingUrl=https://staging.example.com`. Use the staging configuration for physical-device testing.

## Matchmaking behavior

**Cancel search** closes the queue connection and returns to your match settings. The server removes queue membership when it processes the disconnect. If it already created a match, the server can still leave that match pending. The **Match found** screen holds the connection for the next milestone, which adds readiness and gameplay. [Android roadmap](../../docs/client/android-roadmap.md)
