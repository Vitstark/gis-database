package ru.itis.importer.component

import org.slf4j.Logger
import org.slf4j.LoggerFactory
import org.springframework.scheduling.annotation.Async
import org.springframework.stereotype.Component
import org.springframework.transaction.annotation.Propagation
import org.springframework.transaction.annotation.Transactional
import org.springframework.web.client.HttpClientErrorException
import ru.itis.importer.records.CadastralObject
import ru.itis.importer.integration.repository.CadastralRepository
import ru.itis.importer.web.clients.NspdClient
import ru.itis.jooq.`_`.enums.ObjectStatus

@Component
class CadastralComponent(
    private val cadastralRepository: CadastralRepository,
    private val nspdClient: NspdClient
) {
    private val log: Logger = LoggerFactory.getLogger(CadastralComponent::class.java)

    fun getNotActualCadastrals(): List<CadastralObject> {
        return cadastralRepository.getNotActualCadastrals()
    }

    @Transactional
    @Async("cadastralTaskExecutor")
    fun updateCadastralInfo(cadastralObject: CadastralObject) {
        try {
            val cadastralInfo = nspdClient.getCadastralInfo(cadastralObject)
            cadastralRepository.updateCadastralObjectData(cadastralObject, cadastralInfo)
            cadastralRepository.setObjectStatus(cadastralObject.objectCode, ObjectStatus.SUCCESS)
            log.info("Successful update data for cadastral = {}", cadastralObject)
        } catch (e: HttpClientErrorException.NotFound) {
            cadastralRepository.setObjectStatus(cadastralObject.objectCode, ObjectStatus.`NOT FOUND`)
            log.error("Not found data for cadastral = {}", cadastralObject, e)
        } catch (e: Exception) {
            cadastralRepository.setObjectStatus(cadastralObject.objectCode, ObjectStatus.ERROR)
            log.error("Error while update data for cadastral = {}", cadastralObject, e)
        }
    }

    @Transactional(propagation = Propagation.REQUIRES_NEW)
    fun insertCadastral(cadastralNumber: String) {
        try {
            val parsedNumber = parseCadastral(cadastralNumber)
            cadastralRepository.upsertArea(parsedNumber.areaCode, parsedNumber.regionCode)
            cadastralRepository.upsertQuarter(parsedNumber.quarterCode, parsedNumber.areaCode)
            cadastralRepository.upsertObject(parsedNumber.objectCode, parsedNumber.quarterCode)
            log.info("Succssful insert cadastral = {}", cadastralNumber)
        } catch (e: Exception) {
            log.error("Error while insert cadastral = {}", cadastralNumber, e)
        }
    }

    private fun parseCadastral(cadastralNumber: String): CadastralObject {
        val cadastralBits = cadastralNumber.split(":")
        if (cadastralBits.size != 4) throw IllegalArgumentException("Invalid cadastral number: $cadastralNumber")
        return CadastralObject(
            cadastralBits[0].toInt(),
            cadastralBits[1].toInt(),
            cadastralBits[2].toInt(),
            cadastralBits[3].toInt()
        )
    }
}