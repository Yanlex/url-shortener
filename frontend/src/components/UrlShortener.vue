<script setup lang="ts">
import { ref } from "vue";
import { v4 as uuidv4 } from 'uuid';
import axios, { AxiosError} from 'axios';
import { onMounted } from "vue";

interface UrlData {
  base: string;
  shortUrl: string;
  serverReps: string;
}

// interface ApiResponse {
//   shortUrl: string;
// }

interface UserUrls {
  original_url: string
  short_url: string
}

const userUrl = ref("");
const backendURL = "localhost:9462"
var youSendBaseUrl = ref<UrlData[]>([]);

const api = axios.create({
  baseURL: `http://${backendURL}`,
});

api.interceptors.request.use((config) => {
  const userId = getCookie('user_id');
  if (userId) {
    config.headers['X-User-ID'] = userId;
  }
  return config;
});

let userId = getCookie('user_id');

if (!userId) {
  userId = uuidv4();
  if (typeof userId === 'string') {
    document.cookie = `user_id=${userId}; path=/; max-age=31536000`;
  } else {
    console.error('Ошибка генерации user_id');
  }
}

function getCookie(name: string): string | null {
  const cookies = document.cookie.split('; ');
  const cookie = cookies.find(row => row.startsWith(`${name}=`));
  return cookie ? cookie.split('=')[1] : null;
}

async function myFetch(): Promise<void> {
  try {
    // const response: AxiosResponse<ApiResponse> =
    await api.post('/url-short', { baseUrl: userUrl.value, user: userId }, {
      headers: {
        "Content-Type": "application/json",
      },
    });
    youSendBaseUrl = ref<UrlData[]>([])
    getAllUserUrls()
 
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      let errorMessage: string = "Произошла ошибка";

      if (axiosError.response) {
        errorMessage = axiosError.response.data ? JSON.stringify(axiosError.response.data) : axiosError.response.statusText;
      } else if (axiosError.request) {
        errorMessage = "Сервер не ответил";
      } else {
        errorMessage = axiosError.message;
      }
      youSendBaseUrl.value.push({
        base: "",
        shortUrl: "",
        serverReps: errorMessage,
      });

      console.error(axiosError);
    } else {
      console.error("Неизвестная ошибка:", error);
    }
  } finally {
    userUrl.value = "";
  }
}


function urlsFromServer(ne: UserUrls[] ): void {

  for (const item of ne) {

    youSendBaseUrl.value.push({
    base: item.original_url,
    shortUrl: item.short_url,
    serverReps: '',
  });
  }

}
async function getAllUserUrls(): Promise<void> {
  try {
    const response = await api.get<UserUrls[]>('/user-urls', {
    params: {
      user: userId,
    },
    headers: {
      "Content-Type": "application/json",
    },
  });

  urlsFromServer(response.data)

  } catch (error) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      let errorMessage: string = "Произошла ошибка";

      if (axiosError.response) {
        errorMessage = axiosError.response.data ? JSON.stringify(axiosError.response.data) : axiosError.response.statusText;
      } else if (axiosError.request) {
        errorMessage = "Сервер не ответил";
      } else {
        errorMessage = axiosError.message;
      }
      youSendBaseUrl.value.push({
        base: "",
        shortUrl: "",
        serverReps: errorMessage,
      });

      console.error(axiosError);
    } else {
      console.error("Неизвестная ошибка:", error);
    }
  }
    

}

onMounted(() => {
getAllUserUrls()
});
</script>

<template>
  <header class="mt-20">
    <div class="text-6xl italic text-gray-800 sm:text-7xl lg:text-8xl xl:text-9xl">
      <h1 class="">URL</h1>
      <h1>Shortener</h1>
    </div>
  </header>

  <main class="flex flex-col grow">
    <label for="baseUrl" class="block text-sm/6 font-medium text-gray-900">URL</label>
    <div class="flex flex-row w-9/10 max-md:flex-wrap max-md:w-9/10">
      <input
        type="url"
        name="baseUrl"
        id="baseUrl"
        v-model="userUrl"
        @keyup.enter="myFetch"
        class="block grow py-1.5 pr-3 pl-1 border-3 border-cyan-600 text-base text-gray-900 placeholder:text-gray-400 focus:outline-none sm:text-sm/6"
        placeholder="example.com"
      />
      <button
        @click="myFetch"
        class="font-semibold border-cyan-600 py-1.5 pr-3 pl-1.5 border-3 bg-cyan-600 text-blue-50 max-md:w-full max-md:m-auto max-md:mt-1"
      >
        Отправить
      </button>
    </div>
    <span v-for="(url, index) in youSendBaseUrl" :key="index">
      <template v-if="url.base && url.shortUrl">
        {{ `Вы отправили: ${url.base} `}}
        <a  target="_blank" :href="`http://${backendURL}/user-urls/${url.shortUrl}`" class="underline decoration-sky-500 decoration-4 underline-offset-4">
          {{ url.shortUrl }}
    </a>
      </template>
      <template v-if="url.serverReps">
        {{ `${url.serverReps}` }}
      </template>
    </span>
  </main>
  <footer>Я FУТЕР!</footer>
</template>

<style scoped>
.read-the-docs {
  color: #888;
}
</style>