package ru.itis.importer

import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.runApplication

@SpringBootApplication
class ImporterApplication

fun main(args: Array<String>) {
    runApplication<ImporterApplication>(*args)
}
