plugins {
    kotlin("jvm") version "1.9.25"
    kotlin("plugin.spring") version "1.9.25"
    id("org.springframework.boot") version "3.5.7"
    id("io.spring.dependency-management") version "1.1.7"
    id("org.jooq.jooq-codegen-gradle") version "3.19.3"
    id("org.liquibase.gradle") version "2.2.1"
}

group = "ru.itis"
version = "0.0.1-SNAPSHOT"
description = "importer"

java {
    toolchain {
        languageVersion = JavaLanguageVersion.of(17)
    }
}

configurations {
    compileOnly {
        extendsFrom(configurations.annotationProcessor.get())
    }
}

repositories {
    mavenCentral()
}

val changeLogFile: String by project

val databaseDriver: String by project
val databaseUrl: String by project
val databaseUsername: String by project
val databasePassword: String by project

dependencies {
    implementation("org.springframework.boot:spring-boot-starter-actuator")
    implementation("org.springframework.boot:spring-boot-starter-web")
    implementation("org.springframework.boot:spring-boot-starter-jooq")
    implementation("com.fasterxml.jackson.module:jackson-module-kotlin")
    implementation("com.fasterxml.jackson.datatype:jackson-datatype-jsr310")
    implementation("org.jetbrains.kotlin:kotlin-reflect")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.8.1")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-reactor:1.8.1")

    developmentOnly("org.springframework.boot:spring-boot-devtools")

    runtimeOnly("org.postgresql:postgresql")

    annotationProcessor("org.springframework.boot:spring-boot-configuration-processor")
    annotationProcessor("org.projectlombok:lombok")

    testImplementation("org.springframework.boot:spring-boot-starter-test")
    testImplementation("org.jetbrains.kotlin:kotlin-test-junit5")
    testRuntimeOnly("org.junit.platform:junit-platform-launcher")

    jooqCodegen("org.postgresql:postgresql:42.7.2")
    jooqCodegen("org.jooq:jooq-meta-extensions-liquibase:3.19.3")
}

kotlin {
    compilerOptions {
        freeCompilerArgs.addAll("-Xjsr305=strict")
    }
}

tasks.withType<Test> {
    useJUnitPlatform()
}

liquibase {
    activities.register("main") {
        this.arguments = mapOf(
            "changeLogFile" to file(changeLogFile),
            "url" to databaseUrl,
            "username" to databaseUsername,
            "password" to databasePassword
        )
    }
}

jooq {
    configuration {
        jdbc {
            driver = databaseDriver
            url = databaseUrl
            username = databaseUsername
            password = databasePassword
        }
        generator {
            name = "org.jooq.codegen.KotlinGenerator"

            database {
                name = "org.jooq.meta.postgres.PostgresDatabase"
                inputSchema = "public"
                includes = ".*"
                excludes = ".*databasechangelog.*"

                properties {
                    property {
                        key = "dialect"
                        value = "POSTGRESQL"
                    }

                    property {
                        key = "scripts"
                        value = changeLogFile
                    }

                    property {
                        key = "rootPath"
                        value = "$projectDir"
                    }
                }
            }
            target {
                packageName = "ru.itis.jooq."
                directory = "src/main/kotlin"
            }

            generate {
                isPojosAsKotlinDataClasses = true
            }
        }
    }
}
