package ru.itis.importer.web.configuration

import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.springframework.web.client.RestClient

@Configuration
class WebConfiguration {

    @Bean
    fun restClient() = RestClient.builder()
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

}