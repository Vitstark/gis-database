package ru.itis.importer.web.dto

import com.fasterxml.jackson.core.JsonParser
import com.fasterxml.jackson.databind.DeserializationContext
import com.fasterxml.jackson.databind.JsonDeserializer
import com.fasterxml.jackson.databind.JsonNode
import com.fasterxml.jackson.databind.annotation.JsonDeserialize
import java.math.BigDecimal
import java.util.Objects

@JsonDeserialize(using = CadastralObjectInfoDeserializer::class)
class CadastralObjectInfo(
    val rawData: String, // Весь Json ответ

    val area: Int?, // data.features[0].properties.options.specified_area
    val costValue: BigDecimal?, // data.features[0].properties.options.cost_value

    val permittedUseEstablishedByDocument: String?, // data.features[0].properties.options.permitted_use_established_by_document
    val rightType: String?, // data.features[0].properties.options.right_type
    val status: String?, // data.features[0].properties.options.status

    val landRecordType: String?, // data.features[0].properties.options.land_record_type
    val landRecordSubtype: String?, // data.features[0].properties.options.land_record_subtype
    val landRecordCategoryType: String?, // data.features[0].properties.options.land_record_category_type
)

class CadastralObjectInfoDeserializer : JsonDeserializer<CadastralObjectInfo>() {
    override fun deserialize(p: JsonParser, ctxt: DeserializationContext): CadastralObjectInfo? {


        val node: JsonNode = p.codec.readTree(p) ?: return null

        // Сохраняем весь JSON как строку
        val rawData = node.toString()

        // Извлекаем нужные поля из сложной структуры
        val area = node.path("data")
            .path("features")
            .get(0)
            .path("properties")
            .path("options")
            .path("area")
            .asInt()

        val declaredArea = node.path("data")
            .path("features")
            .get(0)
            .path("properties")
            .path("options")
            .path("declared_area")
            .asInt()

        val specified_area = node.path("data")
            .path("features")
            .get(0)
            .path("properties")
            .path("options")
            .path("specified_area")
            .asInt()

        val landRecordCategoryType = node.path("data")
            .path("features")
            .get(0)
            .path("properties")
            .path("options")
            .path("land_record_category_type")
            .asText(null)

        val permittedUseEstablishedByDocument = node.path("data")
            .path("features")
            .get(0)
            .path("properties")
            .path("options")
            .path("permitted_use_established_by_document")
            .asText(null)

        val costValue = node.path("data")
            .path("features")
            .get(0)
            .path("properties")
            .path("options")
            .path("cost_value")
            .asDouble()
            .toBigDecimal()

        val landRecordType = node.path("data")
            .path("features")
            .get(0)
            .path("properties")
            .path("options")
            .path("land_record_type")
            .asText(null)

        val landRecordSubtype = node.path("data")
            .path("features")
            .get(0)
            .path("properties")
            .path("options")
            .path("land_record_subtype")
            .asText(null)

        val rightType = node.path("data")
            .path("features")
            .get(0)
            .path("properties")
            .path("options")
            .path("right_type")
            .asText(null)

        val status = node.path("data")
            .path("features")
            .get(0)
            .path("properties")
            .path("options")
            .path("status")
            .asText(null)

        return CadastralObjectInfo(
            rawData = rawData,

            area = if (area != 0) area else if (declaredArea != 0) declaredArea else specified_area,
            costValue = costValue,

            permittedUseEstablishedByDocument = permittedUseEstablishedByDocument,
            rightType = rightType,
            status = status,

            landRecordType = landRecordType,
            landRecordSubtype = landRecordSubtype,
            landRecordCategoryType = landRecordCategoryType,
            )
    }
}