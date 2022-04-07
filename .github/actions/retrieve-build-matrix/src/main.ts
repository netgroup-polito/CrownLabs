import * as core from '@actions/core'
import * as fs from 'fs'

type Entry = {
  component: string
  optional: boolean
}

type Matrix = {
  component: string[]
  include: Entry[]
}

try {
  const path: string = core.getInput('path', {required: true})
  const filterOptional: boolean = core.getBooleanInput('filterOptional')

  const inputFile = fs.readFileSync(path)
  const entries: Entry[] = JSON.parse(inputFile.toString()).filter(
    (entry: Entry) => !filterOptional || !entry.optional
  )

  const output: Matrix = {
    component: entries.map(obj => obj.component),
    include: entries
  }

  core.setOutput('matrix', JSON.stringify(output))
} catch (error) {
  if (error instanceof Error) core.setFailed(error.message)
}
