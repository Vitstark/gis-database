package ru.itis.importer.service

import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import org.springframework.web.multipart.MultipartFile
import ru.itis.importer.component.CadastralComponent
import java.util.*
import java.util.concurrent.CompletableFuture
import java.util.concurrent.Executors
import java.util.concurrent.Semaphore

@Service
class CadastralService(
    val cadastralComponent: CadastralComponent
) {
    @Transactional
    fun loadCadastral(file: MultipartFile) {
        val scanner = Scanner(file.inputStream)
        while (scanner.hasNextLine()) {
            cadastralComponent.insertCadastral(scanner.nextLine())
        }
    }

    fun updateCadastralInfos() {
        val notActualCadastralNumbers = cadastralComponent.getNotActualCadastrals()
        notActualCadastralNumbers.forEach { cadastral -> cadastralComponent.updateCadastralInfo(cadastral) }
    }
}