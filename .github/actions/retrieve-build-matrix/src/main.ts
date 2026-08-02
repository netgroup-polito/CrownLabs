import {getInput, setOutput} from '@actions/core'
import {readFileSync} from 'fs'

type Entry = {
  component: string
  optional: boolean
}

type Matrix = {
  component: string[]
  include: Entry[]
}

try {
  const path: string = getInput('path', {required: true})
  const filterOptional: boolean = getInput('filterOptional') === 'true'

  const inputFile = readFileSync(path)
  const entries: Entry[] = JSON.parse(inputFile.toString()).filter(
    (entry: Entry) => !filterOptional || !entry.optional
  )

  const output: Matrix = {
    component: entries.map(obj => obj.component),
    include: entries
  }

  setOutput('matrix', JSON.stringify(output))
} catch (error) {
  if (error instanceof Error) setOutput('error', error.message)
}
