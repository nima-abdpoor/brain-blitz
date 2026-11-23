import java.net.URI
import org.jetbrains.kotlin.gradle.dsl.JvmTarget

plugins {
    id("com.android.application") version "8.12.0"
    id("org.jetbrains.kotlin.android") version "2.3.0"
    id("org.jetbrains.kotlin.plugin.compose") version "2.3.0"
    id("org.jetbrains.kotlin.plugin.serialization") version "2.3.0"
}

val stagingUrl = providers.gradleProperty("brainBlitzStagingUrl").orNull?.trimEnd('/')
stagingUrl?.let {
    val uri = URI(it)
    require(uri.scheme == "https" && uri.host != null && uri.path.isEmpty() &&
        uri.userInfo == null && uri.query == null && uri.fragment == null) {
        "brainBlitzStagingUrl must be an HTTPS origin, such as https://staging.example.com"
    }
}

android {
    namespace = "com.brainblitz.android"
    compileSdk = 35

    defaultConfig {
        applicationId = "com.brainblitz.android"
        minSdk = 26
        targetSdk = 35
        versionCode = 1
        versionName = "0.1.0"
    }

    buildTypes {
        debug {
            val baseUrl = stagingUrl ?: "http://10.0.2.2:8088"
            buildConfigField("String", "GATEWAY_URL", "\"$baseUrl\"")
        }
        release {
            buildConfigField("String", "GATEWAY_URL", "\"${stagingUrl.orEmpty()}\"")
        }
    }

    buildFeatures {
        buildConfig = true
        compose = true
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
}

kotlin {
    compilerOptions {
        jvmTarget = JvmTarget.JVM_17
    }
}

dependencies {
    implementation("androidx.activity:activity-compose:1.10.1")
    implementation("androidx.compose.material3:material3:1.3.2")
    implementation("androidx.lifecycle:lifecycle-runtime-compose:2.8.7")
    implementation("androidx.lifecycle:lifecycle-viewmodel-ktx:2.8.7")
    implementation("com.squareup.okhttp3:okhttp:4.12.0")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-android:1.9.0")
    implementation("org.jetbrains.kotlinx:kotlinx-serialization-json:1.9.0")

    testImplementation("junit:junit:4.13.2")
    testImplementation("com.squareup.okhttp3:mockwebserver:4.12.0")
}
