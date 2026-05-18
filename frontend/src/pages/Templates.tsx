import { useState } from 'react'
import { Plus, Edit, Trash2, Copy } from 'lucide-react'
import Card from '@/components/common/Card'
import Button from '@/components/common/Button'
import Input from '@/components/common/Input'
import Modal from '@/components/common/Modal'
import Badge from '@/components/common/Badge'
import type { MessageTemplate } from '@/types'

export default function Templates() {
  const [templates, setTemplates] = useState<MessageTemplate[]>([
    {
      id: '1',
      name: 'Welcome Message',
      message: 'Hi {{name}}, welcome to our service! Your phone number is {{phone}}.',
      variables: ['name', 'phone'],
      createdAt: new Date().toISOString(),
    },
  ])
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingTemplate, setEditingTemplate] = useState<MessageTemplate | null>(null)
  const [formData, setFormData] = useState({ name: '', message: '' })

  const extractVariables = (message: string): string[] => {
    const regex = /{{(\w+)}}/g
    const matches = message.matchAll(regex)
    return Array.from(new Set(Array.from(matches, m => m[1])))
  }

  const handleSave = () => {
    const variables = extractVariables(formData.message)

    if (editingTemplate) {
      setTemplates(templates.map(t =>
        t.id === editingTemplate.id
          ? { ...t, name: formData.name, message: formData.message, variables }
          : t
      ))
    } else {
      const newTemplate: MessageTemplate = {
        id: Date.now().toString(),
        name: formData.name,
        message: formData.message,
        variables,
        createdAt: new Date().toISOString(),
      }
      setTemplates([...templates, newTemplate])
    }

    handleCloseModal()
  }

  const handleEdit = (template: MessageTemplate) => {
    setEditingTemplate(template)
    setFormData({ name: template.name, message: template.message })
    setIsModalOpen(true)
  }

  const handleDelete = (id: string) => {
    if (confirm('Are you sure you want to delete this template?')) {
      setTemplates(templates.filter(t => t.id !== id))
    }
  }

  const handleCopy = (message: string) => {
    navigator.clipboard.writeText(message)
    alert('Template copied to clipboard!')
  }

  const handleCloseModal = () => {
    setIsModalOpen(false)
    setEditingTemplate(null)
    setFormData({ name: '', message: '' })
  }

  const handleOpenModal = () => {
    setEditingTemplate(null)
    setFormData({ name: '', message: '' })
    setIsModalOpen(true)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Message Templates</h1>
          <p className="text-gray-600 mt-1">Create and manage reusable message templates</p>
        </div>
        <Button onClick={handleOpenModal}>
          <Plus className="w-4 h-4 mr-2" />
          New Template
        </Button>
      </div>

      {/* Templates Grid */}
      {templates.length === 0 ? (
        <Card>
          <div className="text-center py-12">
            <FileText className="w-16 h-16 text-gray-400 mx-auto mb-4" />
            <h3 className="text-lg font-semibold text-gray-900 mb-2">No templates yet</h3>
            <p className="text-gray-600 mb-4">Create your first message template</p>
            <Button onClick={handleOpenModal}>
              <Plus className="w-4 h-4 mr-2" />
              Create Template
            </Button>
          </div>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {templates.map((template) => (
            <Card key={template.id} className="hover:shadow-md transition-shadow">
              <div className="space-y-4">
                <div className="flex items-start justify-between">
                  <div>
                    <h3 className="font-semibold text-gray-900">{template.name}</h3>
                    <p className="text-xs text-gray-500 mt-1">
                      Created {new Date(template.createdAt).toLocaleDateString()}
                    </p>
                  </div>
                </div>

                <div className="bg-gray-50 rounded-lg p-3">
                  <p className="text-sm text-gray-700 whitespace-pre-wrap line-clamp-4">
                    {template.message}
                  </p>
                </div>

                {template.variables.length > 0 && (
                  <div className="flex flex-wrap gap-2">
                    {template.variables.map((variable) => (
                      <Badge key={variable} variant="info" size="sm">
                        {'{{' + variable + '}}'}
                      </Badge>
                    ))}
                  </div>
                )}

                <div className="flex space-x-2">
                  <Button
                    size="sm"
                    variant="secondary"
                    onClick={() => handleCopy(template.message)}
                    className="flex-1"
                  >
                    <Copy className="w-4 h-4 mr-1" />
                    Copy
                  </Button>
                  <Button
                    size="sm"
                    variant="secondary"
                    onClick={() => handleEdit(template)}
                  >
                    <Edit className="w-4 h-4" />
                  </Button>
                  <Button
                    size="sm"
                    variant="danger"
                    onClick={() => handleDelete(template.id)}
                  >
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}

      {/* Create/Edit Modal */}
      <Modal
        isOpen={isModalOpen}
        onClose={handleCloseModal}
        title={editingTemplate ? 'Edit Template' : 'Create Template'}
        size="lg"
        footer={
          <>
            <Button variant="secondary" onClick={handleCloseModal}>
              Cancel
            </Button>
            <Button
              onClick={handleSave}
              disabled={!formData.name.trim() || !formData.message.trim()}
            >
              {editingTemplate ? 'Update' : 'Create'}
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="Template Name"
            placeholder="e.g., Welcome Message"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
          />

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Message Content
            </label>
            <textarea
              rows={8}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
              placeholder="Enter your message template here. Use {{variable}} for dynamic content."
              value={formData.message}
              onChange={(e) => setFormData({ ...formData, message: e.target.value })}
            />
            <p className="mt-1 text-xs text-gray-500">
              Use {'{{name}}'}, {'{{phone}}'}, or any custom variable
            </p>
          </div>

          {formData.message && extractVariables(formData.message).length > 0 && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Detected Variables
              </label>
              <div className="flex flex-wrap gap-2">
                {extractVariables(formData.message).map((variable) => (
                  <Badge key={variable} variant="info">
                    {'{{' + variable + '}}'}
                  </Badge>
                ))}
              </div>
            </div>
          )}
        </div>
      </Modal>
    </div>
  )
}
