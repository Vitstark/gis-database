package ru.itis.importer.web.controller

import org.springframework.http.MediaType
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.RequestParam
import org.springframework.web.bind.annotation.RestController
import org.springframework.web.multipart.MultipartFile
import ru.itis.importer.service.CadastralService

@RestController
class CadastralController(
    private val cadastralService: CadastralService
) {
    @PostMapping(path = ["/loadCadastralNumbers"], consumes = [MediaType.MULTIPART_FORM_DATA_VALUE])
    fun loadCadastralObjects(@RequestParam file: MultipartFile) {
        cadastralService.loadCadastral(file)
    }

    @PostMapping(path = ["/updateCadastralInfo"])
    fun updateCadastralInfo() {
        cadastralService.updateCadastralInfos()
    }
}