import { JSONSchemaFaker } from 'json-schema-faker'
import type { JsonValue } from 'type-fest/source/basic'
import type { Module, Verb } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import {
  type MetadataCronJob,
  type MetadataIngress,
  type MetadataSQLQuery,
  type MetadataSubscriber,
  type Ref,
  Visibility,
} from '../../../../protos/xyz/block/ftl/schema/v1/schema_pb'
import { defaultForType } from '../../type.utils'

const basePath = `${window.location.protocol}//${window.location.hostname}:8891/`

export const refString = (ref?: Ref): string => {
  if (!ref) {
    return ''
  }
  return `${ref.module}.${ref.name}`
}

interface JsonMap {
  [key: string]: JsonValue
}

const processJsonValue = (value: JsonValue): JsonValue => {
  if (Array.isArray(value)) {
    return value.map((item) => processJsonValue(item))
  }
  if (typeof value === 'object' && value !== null) {
    const result: JsonMap = {}
    for (const [key, val] of Object.entries(value)) {
      result[key] = processJsonValue(val)
    }

    return result
  }
  if (typeof value === 'string') {
    return ''
  }
  if (typeof value === 'number') {
    return 0
  }
  if (typeof value === 'boolean') {
    return false
  }
  return value
}

// biome-ignore lint/suspicious/noExplicitAny: <explanation>
export const simpleJsonSchema = (verb: Verb): any => {
  if (!verb.jsonRequestSchema) {
    return {}
  }

  let schema = JSON.parse(verb.jsonRequestSchema)

  if (schema.properties && isHttpIngress(verb)) {
    const bodySchema = schema.properties.body
    if (bodySchema) {
      schema = {
        ...bodySchema,
        definitions: schema.definitions,
        required: schema.required.includes('body') ? bodySchema.required : [],
      }
    } else {
      schema = {}
    }
  }

  return schema
}

export const defaultRequest = (verb?: Verb): string => {
  if (!verb) {
    return ''
  }

  if (!verb.jsonRequestSchema) {
    const defaultValue = verb.verb?.request ? defaultForType(verb.verb.request) : null
    return defaultValue !== null ? JSON.stringify(defaultValue, null, 2) : '{}'
  }

  const schema = simpleJsonSchema(verb)
  JSONSchemaFaker.option({
    alwaysFakeOptionals: true,
    optionalsProbability: 0,
    useDefaultValue: false,
    minItems: 0,
    maxItems: 0,
    minLength: 0,
    maxLength: 0,
  })

  JSONSchemaFaker.format('date-time', () => {
    return new Date().toISOString()
  })

  try {
    let fake = JSONSchemaFaker.generate(schema)
    if (fake) {
      fake = processJsonValue(fake)
    }

    return JSON.stringify(fake, null, 2) ?? '{}'
  } catch (error) {
    console.error(error)
    return '{}'
  }
}

export const ingress = (verb?: Verb) => {
  return (verb?.verb?.metadata?.find((meta) => meta.value.case === 'ingress')?.value?.value as MetadataIngress) || null
}

export const cron = (verb?: Verb) => {
  return (verb?.verb?.metadata?.find((meta) => meta.value.case === 'cronJob')?.value?.value as MetadataCronJob) || null
}

export const sqlquery = (verb?: Verb) => {
  return (verb?.verb?.metadata?.find((meta) => meta.value.case === 'sqlQuery')?.value?.value as MetadataSQLQuery) || null
}

export const requestType = (verb?: Verb) => {
  const ingress = (verb?.verb?.metadata?.find((meta) => meta.value.case === 'ingress')?.value?.value as MetadataIngress) || null
  if (ingress) {
    return ingress.method
  }

  const cron = (verb?.verb?.metadata?.find((meta) => meta.value.case === 'cronJob')?.value?.value as MetadataCronJob) || null
  if (cron) {
    return 'CRON'
  }

  const subscriber = (verb?.verb?.metadata?.find((meta) => meta.value.case === 'subscriber')?.value?.value as MetadataSubscriber) || null
  if (subscriber) {
    return 'SUB'
  }

  const sqlquery = (verb?.verb?.metadata?.find((meta) => meta.value.case === 'sqlQuery')?.value?.value as MetadataSQLQuery) || null
  if (sqlquery) {
    return 'SQL'
  }

  return 'CALL'
}

export const httpRequestPath = (verb?: Verb) => {
  const ingress = (verb?.verb?.metadata?.find((meta) => meta.value.case === 'ingress')?.value?.value as MetadataIngress) || null
  if (ingress) {
    return ingress.path
      .map((p) => {
        switch (p.value.case) {
          case 'ingressPathLiteral':
            return p.value.value.text
          case 'ingressPathParameter':
            return `{${p.value.value.name}}`
          default:
            return ''
        }
      })
      .join('/')
  }
}

export const fullRequestPath = (module?: Module, verb?: Verb) => {
  const ingress = (verb?.verb?.metadata?.find((meta) => meta.value.case === 'ingress')?.value?.value as MetadataIngress) || null
  if (ingress) {
    return basePath + httpRequestPath(verb)
  }

  return [module?.name, verb?.verb?.name]
    .filter(Boolean)
    .map((v) => v)
    .join('.')
}

export const httpPopulatedRequestPath = (module?: Module, verb?: Verb) => {
  return fullRequestPath(module, verb).replaceAll(/{([^}]*)}/g, '$1')
}

export const isHttpIngress = (verb?: Verb) => {
  return ingress(verb)?.type === 'http'
}

export const isCron = (verb?: Verb) => {
  return !!cron(verb)
}

export const isExported = (verb?: Verb) => {
  const visibility = verb?.verb?.visibility
  return visibility === Visibility.SCOPE_MODULE || visibility === Visibility.SCOPE_REALM
}

export const createVerbRequest = (path: string, verb?: Verb, editorText?: string, headers?: string) => {
  if (!verb || !editorText) {
    return new Uint8Array()
  }

  let requestJson = JSON.parse(editorText)

  if (isHttpIngress(verb)) {
    // biome-ignore lint/suspicious/noExplicitAny: <explanation>
    const newRequestJson: Record<string, any> = {}
    const httpIngress = ingress(verb)
    if (httpIngress) {
      newRequestJson.method = httpIngress.method
      newRequestJson.path = path.replace(basePath, '')
    }
    newRequestJson.headers = JSON.parse(headers ?? '{}')
    newRequestJson.body = requestJson
    requestJson = newRequestJson
  }

  const textEncoder = new TextEncoder()
  const encoded = textEncoder.encode(JSON.stringify(requestJson))
  return encoded
}

export const generateCliCommand = (verb: Verb, path: string, header: string, body: string) => {
  const method = requestType(verb)
  return method === 'CALL' ? generateFtlCallCommand(path, body) : generateCurlCommand(method, path, header, body)
}

const generateFtlCallCommand = (path: string, editorText: string) => {
  const command = `ftl call ${path} '${editorText}'`
  return command
}

const generateCurlCommand = (method: string, path: string, header: string, body: string) => {
  const headers = JSON.parse(header)

  let curlCommand = `curl -X ${method.toUpperCase()} "${path}"`

  for (const [key, value] of Object.entries(headers)) {
    curlCommand += ` -H "${key}: ${value}"`
  }

  curlCommand += ' -H "Content-Type: application/json"'

  if (method === 'POST' || method === 'PUT') {
    curlCommand += ` -d '${body}'`
  }

  return curlCommand
}
