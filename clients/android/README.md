# BrainBlitz for Android

The first milestone checks the BrainBlitz gateway from an Android device. Sign in with a fresh test account to verify the profile, player counts, question timers, and categories from the initial game WebSocket frame.

## Run locally

1. Open `clients/android` in Android Studio. Use its bundled JDK and install Android SDK Platform 35 and Build Tools 35.0.0 through SDK Manager if needed. The app supports Android 8.0 (API 26) and newer.
2. Start the backend from the repository root with `docker compose up -d --wait`. The default debug build connects to the host gateway at `http://10.0.2.2:8088` on the Android emulator.
3. Create a disposable account with `POST /user-service/public/api/v1/signup` through the gateway, then run the `app` configuration on the emulator and select **Check connection** after entering the account credentials. Success displays the profile name and match settings. [Backend setup](../../docs/run-on-another-computer.md) · [Signup contract](../../README.md#user-service-user-service)

## Build and test

From `clients/android`, with JDK 17 or newer and the Android SDK configured, run:

```sh
./gradlew :app:assembleDebug :app:testDebugUnitTest :app:lintDebug
```

The debug APK is at `app/build/outputs/apk/debug/app-debug.apk`. To build against a staging gateway, set `brainBlitzStagingUrl` to its HTTPS origin in your Gradle properties or pass `-PbrainBlitzStagingUrl=https://staging.example.com`. Use the staging configuration for physical-device testing.

Use a fresh account for the connection check. The current server sends the initial options frame to accounts with no active game or after a game has finished. [Android roadmap](../../docs/client/android-roadmap.md)
