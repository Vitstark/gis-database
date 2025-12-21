package ru.itis.importer.records

data class CadastralObject(
    val regionCode: Int,
    val areaCode: Int,
    val quarterCode: Int,
    val objectCode: Int
) {
    override fun toString(): String {
        return "$regionCode:$areaCode:$quarterCode:$objectCode"
    }
}
