<template>
  <v-data-table
    :headers="headers"
    :items="desserts"
    :search="search"
    :items-per-page=-1
    class="elevation-1"
    item-key="key"
  >
    <template v-slot:top>
      <v-toolbar flat>
        <v-toolbar-title>Service - {{ service }}</v-toolbar-title>
        <v-divider class="mx-4" inset vertical />
        
        <v-spacer />

        <v-dialog persistent v-model="dialog" max-width="800px">
          <template v-slot:activator="{ on, attrs }">
            <v-btn color="primary" dark text class="mb-2" v-bind="attrs" v-on="on">
              New Key
            </v-btn>
            <v-menu offset-y>
              <template v-slot:activator="{ on, attrs }">
                <v-btn color="primary" dark text class="mb-2" v-bind="attrs" v-on="on">Services</v-btn>
              </template>
              <v-list>
                <v-subheader>SELECT SERVICE</v-subheader>
                <v-list-item-group mandatory v-model="name">
                  <v-list-item link v-for="name in services" :key="name" @click="getConfig(name)">
                    <v-list-item-title>{{ name }}</v-list-item-title>
                  </v-list-item>
                </v-list-item-group>
              </v-list>
            </v-menu>
            <v-text-field hide-details v-model="search" placeholder="Search" append-icon="mdi-magnify" />
          </template>
          <v-card>
            <v-card-title>
              <span class="headline" v-if="addingNewKey">Add Key - {{ service }}</span>
              <span class="headline" v-else>Edit Key - {{ service }}</span>
            </v-card-title>

            <v-card-text>
              <v-container>
                <v-row>
                  <v-col cols="12" sm="12" md="8">
                    <v-text-field :readonly="!addingNewKey" v-model="editedItem.key" label="Configuration key" />
                  </v-col>
                  <v-col cols="12" sm="12" md="4">
                    <v-combobox required v-model="editedItem.type" :items='typeSelect' label="Key type" />
                  </v-col>
                  <v-col cols="12" sm="12" md="12">
                    <v-text-field required v-model="editedItem.comment" label="Comment" />
                  </v-col>
                  <v-col cols="12" sm="12" md="12">
                    <v-textarea v-model="editedItem.value" label="Key value" />
                  </v-col>
                </v-row>
              </v-container>
            </v-card-text>

            <v-card-actions>
              <v-spacer />
              <v-btn text color="warning darken-1" @click="close">Cancel</v-btn>
              <v-btn text color="primary darken-1" @click="save">Save</v-btn>
            </v-card-actions>
          </v-card>
        </v-dialog>

        <v-dialog persistent v-model="dialogDelete" max-width="500px">
          <v-card>
            <v-card-title class="headline">Deleting key, are you sure?</v-card-title>
            <v-card-text>
              <v-container>
                <v-alert dense outlined type="error">{{ editedItem.key }}</v-alert>
              </v-container>
            </v-card-text>
            <v-card-actions>
              <v-spacer />
              <v-btn text color="warning darken-1" @click="closeDelete">Cancel</v-btn>
              <v-btn text color="blue darken-1" @click="deleteItemConfirm">OK</v-btn>
            </v-card-actions>
          </v-card>
        </v-dialog>
      </v-toolbar>
    </template>

    <template v-slot:item.actions="{ item }">
      <v-icon small class="mr-2" @click="editItem(item)">mdi-pencil</v-icon>
      <v-icon small @click="deleteItem(item)">mdi-delete</v-icon>
    </template>
  </v-data-table>
</template>

<script>
import axios from 'axios'

export default {
  data: () => ({
    dialog: false, // create/edit dialog
    dialogDelete: false, // delete confirm dialog
    search: '', // search content
    name: 0, // service list name id binding
    service: '', // current service name
    services: [], // service list
    editedIndex: -1, // index of current editing row
    editedItem: {},
    defaultItem: {
      key: '',
      type: '',
      value: '',
      comment: '',
    },
    typeSelect: [
      'bool',
      'int',
      'uint',
      'string',
      'float',
      '[]bool',
      '[]int',
      '[]uint',
      '[]string',
      '[]float',
    ],
    headers: [
      { text: 'Configuration Key', value: 'key' },
      { text: 'Key Type', value: 'type', sortable: false },
      { text: 'Configuration Value', value: 'value', sortable: false },
      { text: 'Comment', value: 'comment', sortable: false },
      { text: 'Actions', value: 'actions', sortable: false },
    ],
    desserts: [],
  }),

  computed: {
    addingNewKey() {
      return this.editedIndex < 0
    }
  },

  created() {
    this.initialize()
  },

  methods: {
    initialize() {
      axios.get('/api/service/').then(res => {
        if (res.data.code !== 0) {
          return
        }
        res.data.data.forEach(item => {
          this.services.push(item)
        })
        if (localStorage.service !== undefined) {
          this.getConfig(localStorage.service)
          this.name = this.services.indexOf(localStorage.service)
        } else {
          this.getConfig(this.services[0])
        }
      }).catch(err => {
        console.log(err)
      })
    },

    refresh(service) {
      axios.get('/api/config/' + service).then(res => {
        if (res.data.code !== 0) {
          console.log(res.data.code)
          return
        }
        localStorage.service = service
        this.service = service
        this.desserts = res.data.data
      }).catch(err => {
        console.log(err)
      })
    },

    getConfig(service) {
      if (service === this.service) {
        return
      }
      this.refresh(service)
    },

    save() {
      let data = JSON.stringify(this.editedItem)
      axios.put('/api/config/' + this.service + '/' + this.editedItem.key, data).then(res => {
        if (res.data.code !== 0) {
          return
        }
        this.refresh(this.service)
      }).catch(err => {
        console.log(err)
      })

      this.close()
    },

    editItem(item) {
      this.editedIndex = this.desserts.indexOf(item)
      this.editedItem = Object.assign({}, item)
      this.dialog = true
    },

    deleteItem(item) {
      this.editedIndex = this.desserts.indexOf(item)
      this.editedItem = Object.assign({}, item)
      this.dialogDelete = true
    },

    deleteItemConfirm() {
      axios.delete('/api/config/' + this.service + '/' + this.editedItem.key).then(res => {
        if (res.data.code !== 0) {
          return
        }
        this.desserts.splice(this.editedIndex, 1)
      }).catch(err => {
        console.log(err)
      })

      this.closeDelete()
    },

    close() {
      this.dialog = false
      this.$nextTick(() => {
        this.editedItem = Object.assign({}, this.defaultItem)
        this.editedIndex = -1
      })
    },

    closeDelete() {
      this.dialogDelete = false
      this.$nextTick(() => {
        this.editedItem = Object.assign({}, this.defaultItem)
        this.editedIndex = -1
      })
    }
  },
}
</script>