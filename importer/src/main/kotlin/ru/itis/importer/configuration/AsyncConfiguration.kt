package ru.itis.importer.configuration

import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.springframework.core.task.AsyncTaskExecutor
import org.springframework.scheduling.annotation.EnableAsync
import org.springframework.scheduling.concurrent.ThreadPoolTaskExecutor
import java.util.concurrent.ThreadPoolExecutor

@Configuration
@EnableAsync
class AsyncConfiguration {

    @Bean("cadastralTaskExecutor")
    fun cadastralTaskExecutor(): AsyncTaskExecutor {
        val executor = ThreadPoolTaskExecutor()
        executor.corePoolSize = 2           // Минимальное количество потоков
        executor.maxPoolSize = 2          // Максимальное количество потоков
        executor.queueCapacity = 0        // Размер очереди
        executor.setThreadNamePrefix("cadastral-")
        executor.setRejectedExecutionHandler(ThreadPoolExecutor.CallerRunsPolicy())
        executor.initialize()
        return executor
    }

}