package ru.itis.importer.integration.repository

import org.jooq.DSLContext
import org.jooq.JSONB
import org.springframework.stereotype.Repository
import org.springframework.transaction.annotation.Propagation
import org.springframework.transaction.annotation.Transactional
import ru.itis.importer.records.CadastralObject
import ru.itis.importer.web.dto.CadastralObjectInfo
import ru.itis.jooq.`_`.enums.ObjectStatus
import ru.itis.jooq.`_`.tables.Area.Companion.AREA
import ru.itis.jooq.`_`.tables.Object.Companion.OBJECT
import ru.itis.jooq.`_`.tables.Quarter.Companion.QUARTER
import ru.itis.jooq.`_`.tables.Region.Companion.REGION
import ru.itis.jooq.`_`.tables.records.AreaRecord
import ru.itis.jooq.`_`.tables.records.ObjectRecord
import ru.itis.jooq.`_`.tables.records.QuarterRecord
import java.time.LocalDate

@Repository
class CadastralRepository(
    val dsl: DSLContext
) {
    fun getNotActualCadastrals(): List<CadastralObject> {
        return dsl.select(
            REGION.CODE,
            AREA.CODE,
            QUARTER.CODE,
            OBJECT.CODE
        )
            .from(REGION)
            .innerJoin(AREA).on(AREA.REGION_CODE.eq(REGION.CODE))
            .innerJoin(QUARTER).on(QUARTER.AREA_CODE.eq(AREA.CODE))
            .innerJoin(OBJECT).on(OBJECT.QUARTER_CODE.eq(QUARTER.CODE))
            .where(OBJECT.LOAD_STATUS.eq(ObjectStatus.NEW))
            .or(OBJECT.UPDATE_DATE.notEqual(LocalDate.now()))
            .fetch()
            .map { CadastralObject(
                it[REGION.CODE]!!.toInt(),
                it[AREA.CODE]!!.toInt(),
                it[QUARTER.CODE]!!,
                it[OBJECT.CODE]!!
            ) }
    }

    fun updateCadastralObjectData(cadastralObject: CadastralObject, info: CadastralObjectInfo) {
        dsl.update(OBJECT)
            .set(OBJECT.UPDATE_DATE, LocalDate.now())
            .set(OBJECT.DATA, JSONB.jsonb(info.rawData))

            .set(OBJECT.AREA, info.area)
            .set(OBJECT.COST_VALUE, info.costValue)

            .set(OBJECT.PERMITTED_USE_ESTABLISHED_BY_DOCUMENT, info.permittedUseEstablishedByDocument)
            .set(OBJECT.RIGHT_TYPE, info.rightType)
            .set(OBJECT.STATUS, info.status)

            .set(OBJECT.LAND_RECORD_TYPE, info.landRecordType)
            .set(OBJECT.LAND_RECORD_SUBTYPE, info.landRecordSubtype)
            .set(OBJECT.LAND_RECORD_CATEGORY_TYPE, info.landRecordCategoryType)

            .where(OBJECT.CODE.eq(cadastralObject.objectCode))
            .execute()
    }

    fun setObjectStatus(objectCode: Int, status: ObjectStatus) {
        dsl.update(OBJECT)
            .set(OBJECT.LOAD_STATUS, status)
            .set(OBJECT.UPDATE_DATE, LocalDate.now())
            .where(OBJECT.CODE.eq(objectCode))
            .execute()
    }

    fun upsertArea(code: Int, regionCode: Int): AreaRecord {
        return dsl.insertInto<AreaRecord>(AREA, AREA.CODE, AREA.REGION_CODE)
            .values(code, regionCode)
            .onConflict(AREA.CODE)
            .doUpdate()
            .set(AREA.REGION_CODE, regionCode.toShort())
            .returning()
            .fetchOne()!!
    }

    fun upsertQuarter(code: Int, areaCode: Int): QuarterRecord {
        return dsl.insertInto<QuarterRecord>(QUARTER, QUARTER.CODE, QUARTER.AREA_CODE)
            .values(code, areaCode)
            .onConflict(QUARTER.CODE)
            .doUpdate()
            .set(QUARTER.AREA_CODE, areaCode.toShort())
            .returning()
            .fetchOne()!!
    }

    fun upsertObject(code: Int, quarterCode: Int, status: ObjectStatus = ObjectStatus.NEW): ObjectRecord {
        return dsl.insertInto<ObjectRecord>(
            OBJECT,
            OBJECT.CODE, OBJECT.QUARTER_CODE, OBJECT.LOAD_STATUS, OBJECT.UPDATE_DATE
        )
            .values(code, quarterCode, status, LocalDate.now())
            .onConflict(OBJECT.CODE)
            .doUpdate()
            .set(OBJECT.QUARTER_CODE, quarterCode)
            .returning()
            .fetchOne()!!
    }

}