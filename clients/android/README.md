# BrainBlitz for Android

Sign in or create a BrainBlitz account, view your profile and statistics, and play multiplayer trivia. The app saves the access token in app-private storage encrypted with a key held by Android Keystore. When the token expires, sign in again. [Android Keystore](https://developer.android.com/privacy-and-security/keystore)

## Run locally

1. Open `clients/android` in Android Studio. Use its bundled JDK and install Android SDK Platform 35 and Build Tools 35.0.0 through SDK Manager if needed. The app supports Android 8.0 (API 26) and newer.
2. Start the current backend from the repository root with `docker compose up -d --build --wait`. Load the demo questions with `docker compose --profile tools run --rm seed-demo`. The default debug build connects to the host gateway at `http://10.0.2.2:8088` on the Android emulator. [Backend setup](../../docs/run-on-another-computer.md)
3. Run the `app` configuration on two emulators and sign in with a distinct account on each. Open **Play**, choose the same category and timer, set the total player count to two, then select **Find match**. Both debug builds should show the same match ID. Select **Ready** on both devices.
4. Answer a question on the first device and check its personal feedback while the second player is still choosing. Have both players answer one round, then let one player miss the next deadline. Check the recap for the correct answer and the missed submission. Rotate a device during a round to check that its deadline and submission are retained.
5. Finish the match, compare both sets of standings, and choose **Play again** to return to the previous match settings. For a mixed-client match, repeat with a Telegram player using a separate BrainBlitz account and the same settings.

## Build and test

From `clients/android`, with JDK 17 or newer and the Android SDK configured, run:

```sh
./gradlew :app:assembleDebug :app:testDebugUnitTest :app:lintDebug
```

The debug APK is at `app/build/outputs/apk/debug/app-debug.apk`. To build against a staging gateway, set `brainBlitzStagingUrl` to its HTTPS origin in your Gradle properties or pass `-PbrainBlitzStagingUrl=https://staging.example.com`. Use the staging configuration for physical-device testing.

## Matchmaking behavior

**Cancel search** waits for the server to confirm you left the queue. If matchmaking has already created a match, the app opens it. Everyone has one minute to select **Ready** by default. After the deadline, expired matches let players search again. [Android roadmap](../../docs/client/android-roadmap.md)

## Test match recovery

- Select **Ready** on the first emulator, close the app process, and relaunch. Check that it returns to the same match and is still waiting for the other player.
- Answer a question on one emulator while the other player waits. Disconnect and reconnect the first emulator, then check that the deadline and accepted answer are preserved.
- Let the other client finish the match while the first app is closed. Relaunch the first app to recover the standings. Returning to the profile or choosing **Play again** dismisses those results on that device.

The app starts a bounded reconnect when the active network changes or a game socket drops, and resumes reconnecting when it returns to the foreground. Match state comes from the authenticated game service. The app asks you to sign in again when the access token expires, then checks for your match.
