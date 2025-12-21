package ru.itis.importer.web.clients

import org.springframework.stereotype.Component
import org.springframework.web.client.RestClient
import org.springframework.web.client.toEntity
import ru.itis.importer.records.CadastralObject
import ru.itis.importer.web.dto.CadastralObjectInfo

@Component
class NspdClient(
    val client: RestClient
) {
    companion object {
        private val INFO_URL = "https://nspd.gov.ru/api/geoportal/v2/search/geoportal"
    }

    fun getCadastralInfo(cadastralObject: CadastralObject): CadastralObjectInfo {
        return client.get()
            .uri(INFO_URL) { uriBuilder ->
                uriBuilder
                    .queryParam("query", cadastralObject.toString())
                    .queryParam("thematicSearchId", "1")
                    .build()
            }
            .retrieve()
            .toEntity<CadastralObjectInfo>()
            .body!!
    }
}

fun main() {
    val result = RestClient.builder()
        .defaultHeaders { headers ->
            // Обязательные заголовки из браузера
            headers.set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")
            headers.set("Accept", "*/*")
            headers.set("Accept-Encoding", "deflate")
            headers.set("Accept-Language", "ru-RU,ru;q=0.8")
            headers.set("Referer", "https://nspd.gov.ru/map?thematic=Default&zoom=19.092349991452554&coordinate_x=5469455.892653173&coordinate_y=7516337.684443036&baseLayerId=235&theme_id=1&is_copy_url=true")

            // Security headers (критически важны!)
            headers.set("sec-ch-ua", "\"Brave\";v=\"143\", \"Chromium\";v=\"143\", \"Not A(Brand\";v=\"24\"")
            headers.set("sec-ch-ua-mobile", "?0")
            headers.set("sec-ch-ua-platform", "\"Windows\"")
            headers.set("sec-fetch-dest", "empty")
            headers.set("sec-fetch-mode", "cors")
            headers.set("sec-fetch-site", "same-origin")
            headers.set("sec-gpc", "1")

            // Дополнительные заголовки
            headers.set("priority", "u=1, i")
            headers.set("connection", "keep-alive")
        }
        .build()
        .get()
        .uri("https://nspd.gov.ru/api/geoportal/v2/search/geoportal") { uriBuilder ->
            uriBuilder
                .queryParam("query", "16:50:011102:413")
                .queryParam("thematicSearchId", "1")
                .build()
        }
        .retrieve()
        .toEntity<CadastralObjectInfo>()

    println(result)
}